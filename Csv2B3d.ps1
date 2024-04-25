go build csv2b3d.go

if (!$?) {
    throw "Failed to compile, aborting"
}

$OutputFolder = "$HOME\Data\GMD\B3D"

if (!(Test-Path $OutputFolder)) {
    New-Item -Path $OutputFolder -ItemType Directory
}

# Remove-Item $OutputFolder\*.b3d
$staticFields = 'zero', 'uniform_1V_km', 'uniform_5V_km', 'uniform_10V_km'

$staticFields | ForEach-Object -ThrottleLimit 6 -Parallel {    
    .\csv2b3d.exe -static -step 0.1 -static -times 1000 "$HOME\Data\GMD\norm_events\$_" "$HOME\Data\GMD\B3D\$_.b3d"
}