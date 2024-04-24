package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// See https://gobyexample.com/command-line-flags for cli parameters

type FieldVector struct {
	lat float64
	lon float64
	Ee  float64
	En  float64
}

type Field []FieldVector

type Point struct {
	lat float64
	lon float64
}

type Vector struct {
	Ee float64
	En float64
}

type CoordRange struct {
	lat0     float64
	lon0     float64
	latStep  float64
	lonStep  float64
	nLat     int
	nLon     int
	nPoints  int
	time0    int
	timeStep int
	nTimes   int
}

func (v Field) Len() int {
	return len(v)
}

func (v Field) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Field) Less(i, j int) bool {
	// return len(s[i]) < len(s[j])
	if v[i].lat < v[j].lat {
		return true
	}

	if math.Abs(v[i].lat-v[j].lat) < 1e-6 {
		return v[i].lon < v[j].lon
	}

	// v[i].lat > v[j].lat
	return false
}

func readLine(line string) (FieldVector, error) {
	s := strings.Split(line, ",")

	lat, err := strconv.ParseFloat(s[0], 64)

	if err != nil {
		return FieldVector{}, errors.New("Error parsing latitude")
	}

	lon, err := strconv.ParseFloat(s[1], 64)

	if err != nil {
		return FieldVector{}, errors.New("Error parsing longitude")
	}

	Ee, err := strconv.ParseFloat(s[2], 64)

	if err != nil {
		return FieldVector{}, errors.New("Error parsing Ee")
		os.Exit(1)
	}

	En, err := strconv.ParseFloat(s[3], 64)

	if err != nil {
		return FieldVector{}, errors.New("Error parsing En")
		os.Exit(1)
	}

	return FieldVector{lat: lat, lon: lon, Ee: Ee, En: En}, nil
}

func readFile(csvPath string) []FieldVector {
	fi, err := os.Open(csvPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %s for reading, aborting\n", csvPath)
		os.Exit(1)
	}

	defer fi.Close()

	scanner := bufio.NewScanner(fi)
	scanner.Scan()

	vectors := []FieldVector{}
	nPoints := 0

	for scanner.Scan() {
		nPoints++
		vec, err := readLine(scanner.Text())

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s:%s: %v", csvPath, nPoints, err)
			os.Exit(1)
		}

		vectors = append(vectors, vec)
	}

	sort.Sort(Field(vectors))
	return vectors
}

func getRange(csvPath string) CoordRange {
	fi, err := os.Open(csvPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %s for reading, aborting\n", csvPath)
		os.Exit(1)
	}

	defer fi.Close()

	scanner := bufio.NewScanner(fi)
	scanner.Scan()

	lats := map[float64]bool{}
	lons := map[float64]bool{}
	nPoints := 0

	for scanner.Scan() {
		nPoints++
		vec, err := readLine(scanner.Text())

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s[%d]: %v", csvPath, nPoints, err)
			os.Exit(1)
		}

		lats[vec.lat] = true
		lons[vec.lon] = true
	}

	lat0 := 0.0
	lon0 := 0.0
	latMax := 0.0
	lonMax := 0.0

	for k, _ := range lats {
		lat0 = k
		latMax = k
		break
	}

	for k, _ := range lons {
		lon0 = k
		lonMax = k
		break
	}

	for k, _ := range lats {
		lat0 = math.Min(lat0, k)
		latMax = math.Max(latMax, k)
	}

	for k, _ := range lons {
		lon0 = math.Min(lon0, k)
		lonMax = math.Max(lonMax, k)
	}

	// fmt.Printf("latMax: %f, nLat: %d\nlonMax: %f, nLon: %d\n", latMax, len(lats), lonMax, len(lons))

	return CoordRange{
		lat0: lat0, lon0: lon0,
		nLat: len(lats), nLon: len(lons),
		latStep: (latMax - lat0) / float64(len(lats)-1), lonStep: (lonMax - lon0) / float64(len(lons)-1),
		nPoints: nPoints,
	}
}

func writeHeader(fo *os.File, cr CoordRange) {
	var magicNumber uint32 = 34280

	if err := binary.Write(fo, binary.LittleEndian, magicNumber); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write magic byte, aborting: %v\n", err)
		os.Exit(1)
	}

	var b3dVersion uint32 = 4

	if err := binary.Write(fo, binary.LittleEndian, b3dVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write B3D version, aborting: %v\n", err)
		os.Exit(1)
	}

	var nMetaStrings uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, nMetaStrings); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write meta string count, aborting: %v\n", err)
		os.Exit(1)
	}

	// var metaString string = "B3D file written by csv2b3d"

	// if err := binary.Write(fo, binary.LittleEndian, metaString); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable to write meta string, aborting: %v\n", err)
	// 	os.Exit(1)
	// }

	// Number of floating point number channels at each point.
	// For data with X and Y directional E-fields, this value will be 2.
	// Convention will be to put X first and then Y.
	var nFloatChannels uint32 = 2

	if err := binary.Write(fo, binary.LittleEndian, nFloatChannels); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write float channel count, aborting: %v\n", err)
		os.Exit(1)
	}

	// Number of byte channels at each point.
	// Usually this value is either zero or one to indicate a quality flag byte
	var nByteChannels uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, nByteChannels); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write byte channel count, aborting: %v\n", err)
		os.Exit(1)
	}

	// Used to indicate the location format. In version 2 this value should be either 0 or 1.
	// If zero the point locations are specified by a grid with the next six FLOAT fields.
	// This was the only approach used in Version 1.  If the LOC_FORMAT is 1 then the points
	// are specified by UNIT number of points and then three location fields for each point.
	var locFormat uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, locFormat); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write location format, aborting: %v\n", err)
		os.Exit(1)
	}

	// Longitude of first point in degrees (only if LOCATION FORMAT = 0)
	lon0 := float32(cr.lon0)

	if err := binary.Write(fo, binary.LittleEndian, lon0); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write longitude origin %f, aborting: %v\n", lon0, err)
		os.Exit(1)
	}

	// Longitude step in degrees (only if LOCATION FORMAT = 0)
	lonStep := float32(cr.lonStep)

	if err := binary.Write(fo, binary.LittleEndian, lonStep); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write longitude step %f, aborting: %v\n", lonStep, err)
		os.Exit(1)
	}

	// Number of longitude points (only if LOCATION FORMAT = 0)
	lonPoints := uint32(cr.nLon)

	if err := binary.Write(fo, binary.LittleEndian, lonPoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write longitude points count %d, aborting: %v\n", lonPoints, err)
		os.Exit(1)
	}

	// Latitude of first point in degrees(only if LOCATION FORMAT = 0)
	lat0 := float32(cr.lat0)

	if err := binary.Write(fo, binary.LittleEndian, lat0); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write latitude origin %f, aborting: %v\n", lat0, err)
		os.Exit(1)
	}

	// Latitude step in degrees(only if LOCATION FORMAT = 0)
	latStep := float32(cr.latStep)

	if err := binary.Write(fo, binary.LittleEndian, latStep); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write latitude step %f, aborting: %v\n", latStep, err)
		os.Exit(1)
	}

	// Number of latitude points(only if LOCATION FORMAT = 0)
	latPoints := uint32(cr.nLat)

	if err := binary.Write(fo, binary.LittleEndian, latPoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write latitude points count %d, aborting: %v\n", latPoints, err)
		os.Exit(1)
	}

	var time0 uint32 = 1462665600

	if err := binary.Write(fo, binary.LittleEndian, time0); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time origin %d, aborting: %v\n", time0, err)
		os.Exit(1)
	}

	// Starting with Version 4.  Indicates the TIME_UNITS scaling used for subsequent time values.
	// Valid entries are 0 indicating milliseconds, 1 indicating seconds, -1 for microseconds,
	// -2 for nanoseconds
	var timeUnits int32 = 1

	if err := binary.Write(fo, binary.LittleEndian, timeUnits); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time units %d, aborting: %v\n", timeUnits, err)
		os.Exit(1)
	}

	// Seconds of first time point, using midnight 1/1/1970 as epoch, not counting leap seconds.
	// (Same as IEEE Std. C37.118.2-2011)
	var timeOffset uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, timeOffset); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time offset %d, aborting: %v\n", timeOffset, err)
		os.Exit(1)
	}

	// Constant time step in TIME_UNITS. If set to zero, indicates variable time step.
	// 10,000 with TIME_UNITS of 0 would be 10 seconds.
	var timeStep uint32 = 100

	if err := binary.Write(fo, binary.LittleEndian, timeStep); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time step %d, aborting: %v\n", timeStep, err)
		os.Exit(1)
	}

	// Number of time points
	timePoints := uint32(cr.nTimes)

	if err := binary.Write(fo, binary.LittleEndian, timePoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time points count %d, aborting: %v\n", timePoints, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: csv2b3d <csvfile> <b3dfile>")
		os.Exit(1)
	}

	csvFolder := os.Args[1]
	b3dFile := os.Args[2]

	fmt.Fprintf(os.Stderr, "In  : %s\nOut: %s\n", csvFolder, b3dFile)

	csvFiles, err := os.ReadDir(csvFolder)
	// csvFiles = csvFiles[0:5]

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to list folder %s, aborting\n", csvFolder)
		os.Exit(1)
	}

	fo, err := os.Create(b3dFile)
	defer fo.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %s for writing, aborting\n", b3dFile)
		os.Exit(1)
	}

	cr := getRange(filepath.Join(csvFolder, csvFiles[0].Name()))
	cr.nTimes = len(csvFiles)

	fmt.Fprintf(os.Stderr, "Lat range: %f:%f:%d\n", cr.lat0, cr.latStep, cr.nLat)
	fmt.Fprintf(os.Stderr, "Lon range: %f:%f:%d\n", cr.lon0, cr.lonStep, cr.nLon)
	fmt.Fprintf(os.Stderr, "Points: %d\n", cr.nPoints)
	fmt.Fprintf(os.Stderr, "Times: %d\n", cr.nTimes)

	writeHeader(fo, cr)

	for i, csvFile := range csvFiles {
		csvPath := filepath.Join(csvFolder, csvFile.Name())
		fmt.Printf("%d: %s\n", i+1, csvFile.Name())
		field := readFile(csvPath)

		for _, vec := range field {
			Ee := float32(vec.Ee)
			En := float32(vec.En)

			if err := binary.Write(fo, binary.LittleEndian, Ee); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to write Ee %f, aborting: %v\n", Ee, err)
				os.Exit(1)
			}

			if err := binary.Write(fo, binary.LittleEndian, En); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to write En %f, aborting: %v\n", En, err)
				os.Exit(1)
			}
		}

	}

}
