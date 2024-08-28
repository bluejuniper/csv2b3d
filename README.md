# CSV2B3D

CSV to B3D V4 converter written in Go. B3D is a binary format
for Powerworld (TM) Simulator used to store time- and 
spatially-varying geoelectric field data. 

For more information on the B3D file format, refer to the
PowerWorld (TM) documentation:

https://www.powerworld.com/knowledge-base/b3d-file-format

## Running

To Build: `go build csv2b3d.go`
 
To Run: `csv2b3d <csvdir> <b3dfile>`

### Options

To set a message in the header: `csv2b3d -message "<message>" <csvdir> <b3dfile>`

To set an optional time step in seconds: `csv2b3d -step <time_step> <csvdir> <b3dfile>`. Note that the default time step is 60 s.

To only run on the first n files: `csv2b3d -times <count> <csvdir> <b3dfile>`

For help: `csv2b3d -help`

Batch Conversion: For an example of a PowerShell 7.x script to process a folder
of folders of input CSV files in parallel, see `Csv2B3d.ps1`

## Expected CSV Format

The input CSV files are expected to have a header line
and data lines in the following order:

1. Latitude (-180 to +180 degrees)
2. Longitude (-180 to +180 degrees)
3. West-East component of electric field
4. South-North component of electric field

Additonally, the CSV files should be named 
with leading zeros such that a string sort will 
return the correct order, e.g.

```
input01.csv
input02.csv
input03.csv
input04.csv
input05.csv
input06.csv
input07.csv
input08.csv
input09.csv
input10.csv
```

Example input file:

``` csv
Lat(Deg),Lon(Deg),Ee(V/km),En(V/km)
0.0,0.0,-0.9956,0.9267
0.0,1.0,-0.1526,0.7598
0.0,2.0,-0.5288,0.3514
0.0,3.0,-0.0806,0.2324
0.0,4.0,-0.5321,0.2654
0.0,5.0,-0.7197,0.3807
0.0,6.0,-0.8127,0.3258
```

## License

This code is provided under a [BSD license](https://github.com/lanl-ansi/PowerModelsGMD.jl/blob/master/LICENSE.md) as part of the Multi-Infrastructure Control and Optimization Toolkit (MICOT) project, LA-CC-13-108.
