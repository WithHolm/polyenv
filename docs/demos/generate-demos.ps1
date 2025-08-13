[cmdletBinding(SupportsShouldProcess)]
param()
$root = git rev-parse --show-toplevel
$dist = join-path $root "dist"
$checksumsFile = join-path $PSScriptRoot "checksums.json"
if (!(test-path $checksumsFile)) {
    $checksums = @{}
}
else {
    $checksums = Get-Content $checksumsFile | ConvertFrom-Json -AsHashtable
}
$checksums = [hashtable]::Synchronized($checksums)

if (!(test-path $dist)) {
    make build
}

$exedir = gci $dist -Filter "polyenv.exe" -Recurse -File
set-alias polyenv $exedir

Get-ChildItem -Path $PSScriptRoot -Filter *.tape -Recurse | ForEach-Object -Parallel {
    $file = $_
    $checksums = $using:checksums
    $dist = $using:dist
    $exedir = $using:exedir
    $PSScriptRoot = $using:PSScriptRoot
    $gifPath = join-path $PSScriptRoot "$($file.BaseName).gif"

    #see if i should skip this file because its already processed
    if (!$checksums.ContainsKey($file.Name)) {
        $checksums[$file.Name] = (Get-FileHash $file.FullName -Algorithm SHA256).Hash
    }
    elseif ($checksums[$file.Name] -eq (Get-FileHash $file.FullName -Algorithm SHA256).Hash) {
        Write-Host "Skipping $file"
        continue
    }

    try {
        $tempDir = join-path ([System.IO.Path]::GetTempPath()) "$($file.BaseName)-temp"
        New-Item -Path $tempDir -ItemType Directory -Force | Out-Null

        Copy-item -Path $file.FullName -Destination $tempDir
        $tempFile = join-path $tempDir $file.Name

        $content = Get-content $file
        $Theming = @(
            "#set theme and style.."
            "Output '$gifPath'"
            "Set Theme 'Catppuccin Frappe'"
            "Set FontSize 30"
            "Set Width 1200"
            "Set Height 600"
        )

        $setup = @(
            "Set Shell pwsh"
            "Hide"
            "Type Set-Alias polyenv '$exedir'"
            "Enter"
            "Type Set-Location '$tempDir'"
            "Enter"
            "Type cls"
            "Enter"
            "Show"
        )

        $Demo = $Theming + $setup + @("#START DEMO", "") + $content
        # New-Item -Path $temp -ItemType File -Force|Out-Null
        # Start-Sleep -Milliseconds 50
        $Demo | Out-File -FilePath $tempFile  -Force
        vhs $tempFile
    }
    finally {
        $tempDir
        get-item $tempDir | Remove-Item -Recurse -Force
        # get-item $temp|Remove-Item
    }
}

$checksums | ConvertTo-Json -Depth 100 | Out-File $checksumsFile