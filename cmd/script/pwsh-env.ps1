$json = & "%s" %s env -o json|Convertfrom-json;
foreach ($key in $json.PSObject.Properties.Name) {
    Write-Host "setting env:$key";
    Set-Item -Path "env:$key" -Value $json.$key -whatif:$false;
}
