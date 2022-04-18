package main

import (
	"fmt"
	// "io"
	"bufio"
	"errors"
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

	for k := range lats {
		lat0 = k
		latMax = k
		break
	}

	for k := range lons {
		lon0 = k
		lonMax = k
		break
	}

	for k := range lats {
		if k < lat0 {
			lat0 = k
		}

		if k > latMax {
			latMax = k
		}

		break
	}

	for k := range lons {
		if k < lon0 {
			lon0 = k
		}

		if k > lonMax {
			lonMax = k
		}

		break
	}

	fmt.Printf("latMax: %f, nLat: %d\nlonMax: %f, nLon: %d\n", latMax, len(lats), lonMax, len(lons))

	return CoordRange{
		lat0: lat0, lon0: lon0,
		nLat: len(lats), nLon: len(lons),
		latStep: (latMax - lat0) / float64(len(lats)), lonStep: (lonMax - lon0) / float64(len(lons)),
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

			fmt.Fprintf(fo, "%f,%f,%f,%f\n", vec.lat, vec.lon, vec.Ee, vec.En)

		}
	}

}
