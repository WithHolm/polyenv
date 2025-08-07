[cmdletbinding()]
# param(
#     [string]$exe,
#     [string]$env
# )

$json = & "%s" %s env -o json;

$jsonObj = ConvertFrom-Json $json;
$jsonObj

# foreach ($key in $jsonObj.PSObject.Properties.Name) {
#     $value = $jsonObj.$key
#     Write-Host "setting env:$key"
#     Set-Item -Path "env:$key" -Value $value -whatif:$false
# }
