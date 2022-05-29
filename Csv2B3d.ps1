go build csv2b3d.go
$BaseFolder = "$HOME\Data\GMD\norm_events"
del data\*.b3d

dir $BaseFolder | ForEach-Object -ThrottleLimit 8 -Parallel {    
    .\csv2b3d.exe $_.FullName "data\$($_.BaseName).b3d"
}