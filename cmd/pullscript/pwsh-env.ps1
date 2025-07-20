[cmdletbinding()]
param(
    [string]$exe,
    [string]$dotenv
)

$json = & $path pull --path $dotenv --out json

$jsonObj = ConvertFrom-Json $json

foreach ($key in $jsonObj.PSObject.Properties.Name) {
    $value = $jsonObj.$key
    Write-Verbose "setting env:$key"
    Set-Item -Path "env:$key" -Value $value -whatif:$false
}
