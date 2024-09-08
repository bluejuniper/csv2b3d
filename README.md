# CSV2B3D

CSV to B3D V4 converter written in Go. B3D is a binary format
for Powerworld (TM) Simulator used to store time- and 
spatially-varying geoelectric field data. This assumes that
latitude and longitude points are set on a uniform grid ranging
between -180 and + 180 degrees. Non-uniform grids are
currently not supported.

For more information on the B3D file format, refer to the
PowerWorld (TM) documentation:

https://www.powerworld.com/knowledge-base/b3d-file-format

## Running

To Build: `go build csv2b3d.go`
 
To Run: `csv2b3d <efield.csv>`

This will generate an output file `efield.b3d`. 

To specify the output file: `csv2b3d <efield.csv> <efield.b3d>`

To specify a message in the b3d header: `csv2b3d -m "message" <efield.csv>`


## Expected CSV Format

The input CSV file is expected to have a header line
and data lines in the following order:

1. Date in YYYY-MM-DD format
2. Time in hh:mm:ss.sss format
3. West-East component of electric field
4. South-North component of electric field
5. Latitude (-180 to +180 degrees)
6. Longitude (-180 to +180 degrees)

## License

This code is provided under a [BSD license](https://github.com/lanl-ansi/PowerModelsGMD.jl/blob/master/LICENSE.md) as part of the Multi-Infrastructure Control and Optimization Toolkit (MICOT) project, LA-CC-13-108.
