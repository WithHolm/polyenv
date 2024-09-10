param(
    [string]$targetOS = "windows-amd64,linux-amd64,linux-ppc64le,darwin-amd64,darwin-arm64",
    [string]$path = "./build"
)

$plaforms = $targetOS -split ","
$jobs = @()
foreach ($platform in $plaforms) {
    Write-Host "building for $platform"
    $job = Start-Job -Name "build-$platform" -ScriptBlock {
        param($platform, $path)
        $env:GOOS = $platform.split("-")[0]
        $env:GOARCH = $platform.split("-")[1]
        $ext = if ($env:GOOS -eq "windows") { ".exe" } else { "" }
        go build -o "./build/$platform/dotenv-myvault$ext" main.go
    } -ArgumentList $platform, $path
    $jobs += $job
}

while($jobs.state -contains 'Running') {
    $nc= $jobs | Where-Object { $_.state -eq 'Running' }
    Write-Host "waiting for $($nc.Count) jobs to finish ($($nc.Name -join ", "))"
    Start-Sleep -Seconds 5
}
$jobs | Wait-Job | Receive-Job