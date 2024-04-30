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
    # .\csv2b3d.exe -step 0.1 -times 1000 "$HOME\Data\GMD\norm_events\$_" "$HOME\Data\GMD\B3D\$_.b3d"
    .\csv2b3d.exe -step 10 -times 10 "$HOME\Data\GMD\norm_events\$_" "$HOME\Data\GMD\B3D\$($_)_dt_10s_tf_100s.b3d"
}

# $DynamicFields = 'blake_20031120_3d', 'blake_20050515_3d', 'blake_scaledA1_3d', 'blake_scaledA2_3d', 'blake_scaledB1_3d', 'blake_scaledB4_3d', 'DTWSept_high'

# $DynamicFields | ForEach-Object -ThrottleLimit 6 -Parallel {    
#     .\csv2b3d.exe -step 60.0 "$HOME\Data\GMD\norm_events\$_" "$HOME\Data\GMD\B3D\$_.b3d"
# }
