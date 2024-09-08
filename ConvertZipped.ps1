param (
    [string]$CsvFile,
    [string]$OutFolder
)

# $dt = Get-Date Format "yyyyMMddTHHmmssffff"
$dt = Get-Date -Format "yyyy-MM-ddTHH-mm-ss"

$n = $CsvFile.Length
$CsvName = $CsvFile.Substring(0, $n-4)

$B3dFile = "$CsvName.b3d"
$ZipFile = "$CsvName-$dt.zip"

$B3dFolder = $B3dFile.Folderectory

if (!(Test-Path $B3dFolder)) {
    mkFolder $B3dFolder
}

if ($OutFolder) {
    $CsvName = $CsvFile.BaseName
    $B3dFile = Join-Path $OutFolder "$CsvName.b3d"
    $ZipFile = Join-Path $OutFolder "$CsvName-$dt.zip"
}

./csv2b3d $CsvFile $B3dFile
# Write-Host -ForegroundColor Green "Wrote to $B3dFile"
Compress-Archive $B3dFile $ZipFile
# Write-Host -ForegroundColor Green "Wrote to $ZipFile"