package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FieldVector struct {
	lat float64
	lon float64
	Ee  float64
	En  float64
}

type CoordRange struct {
	lat0    float64
	lon0    float64
	latStep float64
	lonStep float64
	nLat    int
	nLon    int
	nPoints int
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
			fmt.Fprintf(os.Stderr, "Error reading %s:%s: %v", csvPath, nPoints, err)
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

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: csv2b3d <csvfile> <b3dfile>")
		os.Exit(1)
	}

	csvFolder := os.Args[1]
	b3dFile := os.Args[2]

	fmt.Fprintf(os.Stderr, "In  : %s\nOut: %s\n", csvFolder, b3dFile)

	csvFiles, err := os.ReadDir(csvFolder)

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
	fmt.Fprintf(os.Stderr, "%f:%f:%d\n", cr.lat0, cr.latStep, cr.nLat)
	fmt.Fprintf(os.Stderr, "%f:%f:%d\n", cr.lon0, cr.lonStep, cr.nLon)
	fmt.Fprintf(os.Stderr, "%d\n", cr.nPoints)
	var magicNumber uint32 = 34280

	if err := binary.Write(fo, binary.LittleEndian, magicNumber); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write magic byte to %s, aborting: %v\n", b3dFile, err)
		os.Exit(1)
	}

	var b3dVersion uint32 = 4

	if err := binary.Write(fo, binary.LittleEndian, b3dVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write B3D version to %s, aborting: %v\n", b3dFile, err)
		os.Exit(1)
	}

	var nMetaStrings uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, nMetaStrings); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write meta string count to %s, aborting: %v\n", b3dFile, err)
		os.Exit(1)
	}

	// var metaString string = "B3D file written by csv2b3d"

	// if err := binary.Write(fo, binary.LittleEndian, metaString); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable to write meta string to %s, aborting: %v\n", b3dFile, err)
	// 	os.Exit(1)
	// }

	var nFloatChannels uint32 = 2

	if err := binary.Write(fo, binary.LittleEndian, nFloatChannels); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write float channel count to %s, aborting: %v\n", b3dFile, err)
		os.Exit(1)
	}

	var nByteChannels uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, nByteChannels); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write byte channel count to %s, aborting: %v\n", b3dFile, err)
		os.Exit(1)
	}

	var locFormat uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, locFormat); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write location format to %s, aborting: %v\n", b3dFile, err)
		os.Exit(1)
	}

	lon0 := float32(cr.lon0)

	if err := binary.Write(fo, binary.LittleEndian, lon0); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write longitude origin %f to %s, aborting: %v\n", lon0, b3dFile, err)
		os.Exit(1)
	}

	lonStep := float32(cr.lonStep)

	if err := binary.Write(fo, binary.LittleEndian, lonStep); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write longitude step %f to %s, aborting: %v\n", lonStep, b3dFile, err)
		os.Exit(1)
	}

	lonPoints := uint32(cr.nLon)

	if err := binary.Write(fo, binary.LittleEndian, lonPoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write longitude points count %d to %s, aborting: %v\n", lonPoints, b3dFile, err)
		os.Exit(1)
	}

	lat0 := float32(cr.lat0)

	if err := binary.Write(fo, binary.LittleEndian, lat0); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write latitude origin %f to %s, aborting: %v\n", lat0, b3dFile, err)
		os.Exit(1)
	}

	latStep := float32(cr.latStep)

	if err := binary.Write(fo, binary.LittleEndian, latStep); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write latitude step %f to %s, aborting: %v\n", latStep, b3dFile, err)
		os.Exit(1)
	}

	latPoints := uint32(cr.nLat)

	if err := binary.Write(fo, binary.LittleEndian, latPoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write latitude points count %d to %s, aborting: %v\n", latPoints, b3dFile, err)
		os.Exit(1)
	}

	numPoints := uint32(cr.nPoints)

	if err := binary.Write(fo, binary.LittleEndian, numPoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write latitude points count %d to %s, aborting: %v\n", numPoints, b3dFile, err)
		os.Exit(1)
	}

	var time0 uint32 = 1462665600

	if err := binary.Write(fo, binary.LittleEndian, time0); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time origin %d to %s, aborting: %v\n", time0, b3dFile, err)
		os.Exit(1)
	}

	var timeUnits int32 = 1

	if err := binary.Write(fo, binary.LittleEndian, timeUnits); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time units %d to %s, aborting: %v\n", timeUnits, b3dFile, err)
		os.Exit(1)
	}

	var timeStep uint32 = 60

	if err := binary.Write(fo, binary.LittleEndian, timeStep); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time step %d to %s, aborting: %v\n", timeStep, b3dFile, err)
		os.Exit(1)
	}

	var timeOffset uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, timeOffset); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time offset %d to %s, aborting: %v\n", timeOffset, b3dFile, err)
		os.Exit(1)
	}

	timePoints := uint32(len(csvFiles))

	if err := binary.Write(fo, binary.LittleEndian, timePoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time points count %d to %s, aborting: %v\n", timePoints, b3dFile, err)
		os.Exit(1)
	}

	for i, csvFile := range csvFiles {
		if i >= 3 {
			break
		}

		csvPath := filepath.Join(csvFolder, csvFile.Name())
		fmt.Printf("%d: %s\n", i+1, csvFile.Name())

		fi, err := os.Open(csvPath)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to open %s for reading, aborting\n", csvPath)
			os.Exit(1)
		}

		defer fi.Close()

		scanner := bufio.NewScanner(fi)
		scanner.Scan()

		j := 0

		for scanner.Scan() {
			j++

			if j >= 5 {
				break
			}

			// fmt.Println(scanner.Text())
			// Lat(Deg),Lon(Deg),Ee(V/km),En(V/km),Em(V/km)
			vec, err := readLine(scanner.Text())

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading %s:%s: %v", csvFile, j, err)
				os.Exit(1)
			}

			fmt.Printf("%f,%f,%f,%f\n", vec.lat, vec.lon, vec.Ee, vec.En)
		}
	}

}
