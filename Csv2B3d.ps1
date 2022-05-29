go build csv2b3d.go

$InputFolder = "$HOME\Data\GMD\norm_events"
$OutputFolder = "$HOME\Data\GMD\Events\B3D"

if (!(Test-Path $OutputFolder)) {
    New-Item -Path $OutputFolder -ItemType Directory
}

Remove-Item $OutputFolder\*.b3d

Get-ChildItem $InputFolder | ForEach-Object -ThrottleLimit 8 -Parallel {    
    .\csv2b3d.exe $_.FullName "data\$($_.BaseName).b3d"
}