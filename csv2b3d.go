package main

import (
	"fmt"
	// "io"
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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

	for i, csvFile := range csvFiles {
		if i >= 3 {
			break
		}

		csvPath := filepath.Join(csvFolder, csvFile.Name())
		fmt.Println(csvFile.Name())

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

			fmt.Println(scanner.Text())
			// Lat(Deg),Lon(Deg),Ee(V/km),En(V/km),Em(V/km)
			s := strings.Split(scanner.Text(), ",")

			lat, err := strconv.ParseFloat(s[0], 64)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing latitude on %s:%s, aborting\n", csvPath, j)
				os.Exit(1)
			}

			lon, err := strconv.ParseFloat(s[1], 64)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing longitude on %s:%s, aborting\n", csvPath, j)
				os.Exit(1)
			}

			Ee, err := strconv.ParseFloat(s[2], 64)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing Ee on %s:%s, aborting\n", csvPath, j)
				os.Exit(1)
			}

			En, err := strconv.ParseFloat(s[3], 64)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing En on 1:%s, aborting\n", csvPath, j)
				os.Exit(1)
			}

			fmt.Fprintf(fo, "%f,%f,%f,%f\n", lat, lon, Ee, En)

		}
	}

}
