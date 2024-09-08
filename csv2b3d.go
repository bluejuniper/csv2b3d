package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
    "encoding/binary"
)

// See https://gobyexample.com/command-line-flags for cli parameters
const (
	nsTimeUnits = -2
	usTimeUnits = -1
	msTimeUnits = 0
	sTimeUnits  = 1
)

const (
	measurement_station_location         float64 = 0.0
	measurement_station_location_unknown float64 = -1.0
)

// Date,Time,Ex,Ey,Latitude,Longitude
type FieldVector struct {
	t   float64
	lon float64
	lat float64
	Ee  float64
	En  float64
}

type Location struct {
    lon float64
    lat float64
}

type Field []FieldVector

func (v Field) Len() int {
	return len(v)
}

func (v Field) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// 2024-08-24 13:24:00 (-106.5, 34.0)
// 2024-08-24 13:20:00 (-106.0, 34.5)
func (v Field) Less(i, j int) bool {
	if v[i].t != v[j].t {
		return v[i].t < v[j].t
	}

	if v[i].lat != v[j].lat {
		return v[i].lat < v[j].lat
	}

    return v[i].lon < v[j].lon 	
}

func printField(field Field) {
    fmt.Printf("Step,Time,Longitude,Latitude,Ee,En\n")
    for i, u := range field {
        //fmt.Printf("step: %d, time: %0.3f, lon: %0.3f, lat: %0.3f, x: %0.3f, y: %0.3f\n", i+1, u.t, u.lon, u.lat, u.Ee, u.En)
        fmt.Printf("%d,%0.3f,%0.3f,%0.3f,%0.3f,%0.3f\n", i+1, u.t, u.lon, u.lat, u.Ee, u.En)
    }
}

func printTimes(field Field) {
    times := getTimes(field)

    for i, t := range times {
        fmt.Printf("Step %d, Time: %0.3f\n", i+1, t)
    }
}

func printLocs(field Field) {
    locs := getLocations(field)

    for i, u := range locs {
        fmt.Printf("Step %d, Lon: %0.6f, Lat: %0.6f\n", i+1, u.lon, u.lat)
    }
}


func getTimes(f Field) []float64 {
    // call this if we already know size of t 
    // times := make([]float64, nt)
    times := []float64{}
    tp := f[0].t
    times = append(times, tp)

    for _, v := range f[1:] {
        if v.t > tp {
            tp = v.t
            times = append(times, tp)
        }
    }
 
    return times
}

func getLocations(v Field) []Location {
    locs := []Location{}
    tp := v[0].t
    locs = append(locs, Location{lon: v[0].lon, lat: v[0].lat})

    for _, u := range v[1:] {
        if u.t > tp {
            break
        }
        locs = append(locs, Location{lon: u.lon, lat: u.lat})
    }
    
    return locs
}


func readLine(line string) (FieldVector, error) {
	s := strings.Split(line, ",")

	// 2024-05-12,00:00:30.000
	dts := s[0] + "T" + s[1] + "Z"
	dt, err := time.Parse(time.RFC3339, dts)

	if err != nil {
        return FieldVector{}, errors.New("Error parsing time: " + dts + "\n")
		os.Exit(1)
	}

	t := float64(dt.Unix())

	Ee, err := strconv.ParseFloat(s[2], 64)

	if err != nil {
		return FieldVector{}, errors.New("Error parsing Ee\n")
		os.Exit(1)
	}

	En, err := strconv.ParseFloat(s[3], 64)

	if err != nil {
		return FieldVector{}, errors.New("Error parsing En\n")
		os.Exit(1)
	}

	lat, err := strconv.ParseFloat(s[4], 64)

	if err != nil {
		return FieldVector{}, errors.New("Error parsing latitude\n")
	}

	lon, err := strconv.ParseFloat(s[5], 64)

	if err != nil {
		return FieldVector{}, errors.New("Error parsing longitude\n")
	}

	return FieldVector{t: t, lat: lat, lon: lon, Ee: Ee, En: En}, nil
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


	// Don't need to sort for location format 1
    for i := 1; i < len(vectors); i++ {
        //# field is unsorted
        if Field(vectors).Less(i, i - 1) {
	        sort.Sort(Field(vectors))
            return vectors
        }
    }

	return vectors
}


func writeHeader(fo *os.File, field []FieldVector, message string) {
    const magicNumber uint32 = 34280

	if err := binary.Write(fo, binary.LittleEndian, magicNumber); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write magic byte, aborting: %v\n", err)
		os.Exit(1)
	}

	const b3dVersion uint32 = 4

	if err := binary.Write(fo, binary.LittleEndian, b3dVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write B3D version, aborting: %v\n", err)
		os.Exit(1)
	}

	var nMetaStrings uint32 = 0

	if len(message) > 0 {
		nMetaStrings = 1
	}

	if err := binary.Write(fo, binary.LittleEndian, nMetaStrings); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write meta string count, aborting: %v\n", err)
		os.Exit(1)
	}

	if len(message) > 0 {
		for _, char := range message {
			if err := binary.Write(fo, binary.LittleEndian, char); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to write meta string, aborting: %v\n", err)
				os.Exit(1)
			}
		}

		// https://stackoverflow.com/questions/38007361/how-to-create-a-null-terminated-string-in-go
		if err := binary.Write(fo, binary.LittleEndian, rune(0)); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to write meta string terminator, aborting: %v\n", err)
			os.Exit(1)
		}
	}

	// Number of floating point number channels at each point.
	// For data with X and Y directional E-fields, this value will be 2.
	// Convention will be to put X first and then Y.
	const nFloatChannels uint32 = 2

	if err := binary.Write(fo, binary.LittleEndian, nFloatChannels); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write float channel count, aborting: %v\n", err)
		os.Exit(1)
	}

	// Number of byte channels at each point.
	// Usually this value is either zero or one to indicate a quality flag byte
	const nByteChannels uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, nByteChannels); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write byte channel count, aborting: %v\n", err)
		os.Exit(1)
	}

	// Used to indicate the location format. In version 2 this value should be either 0 or 1.
	// If zero the point locations are specified by a grid with the next six FLOAT fields.
	// This was the only approach used in Version 1.  If the LOC_FORMAT is 1 then the points
	// are specified by UNIT number of points and then three location fields for each point.
	const locFormat uint32 = 1

	if err := binary.Write(fo, binary.LittleEndian, locFormat); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write location format, aborting: %v\n", err)
		os.Exit(1)
	}



	// Number of latitude points(only if LOCATION FORMAT = 0)
    locs := getLocations(field)
    fmt.Fprintf(os.Stderr, "Number of coordinates: %d\n", len(locs))
	nPoints := uint32(len(locs))

	if err := binary.Write(fo, binary.LittleEndian, nPoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write points count %d, aborting: %v\n", nPoints, err)
		os.Exit(1)
	}

	for _, vec := range locs {
		lon := float64(vec.lon)
		lat := float64(vec.lat)
		dist_to_measurement_station := float64(measurement_station_location)

		if err := binary.Write(fo, binary.LittleEndian, lon); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to write longitude %f, aborting: %v\n", lon, err)
			os.Exit(1)
		}

		if err := binary.Write(fo, binary.LittleEndian, lat); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to write latitude %f, aborting: %v\n", lat, err)
			os.Exit(1)
		}

		if err := binary.Write(fo, binary.LittleEndian, dist_to_measurement_station); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to write distance to measurement station %f, aborting: %v\n", dist_to_measurement_station, err)
			os.Exit(1)
		}

	}

	// Seconds of first time point, using midnight 1/1/1970 as epoch, not counting leap seconds.
	// (Same as IEEE Std. C37.118.2-2011)
    times := getTimes(field)
	startTime := uint32(times[0])

	if err := binary.Write(fo, binary.LittleEndian, startTime); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time origin %d, aborting: %v\n", startTime, err)
		os.Exit(1)
	}

	// Starting with Version 4.  Indicates the TIME_UNITS scaling used for subsequent time values.
	// Valid entries are 0 indicating milliseconds, 1 indicating seconds, -1 for microseconds,
	// -2 for nanoseconds
    // Use units of seconds for now 
    // TODO: select smaller time unit based on smallest nonzero timestep
	var timeUnits int32 = sTimeUnits

	//if tr.timeStep >= 1.0 {
	//	timeUnits = sTimeUnits
	//} else if tr.timeStep >= 1e-3 {
	//	timeUnits = msTimeUnits
	//} else if tr.timeStep >= 1e-6 {
	//	timeUnits = usTimeUnits
	//}

	if err := binary.Write(fo, binary.LittleEndian, timeUnits); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time units %d, aborting: %v\n", timeUnits, err)
		os.Exit(1)
	}

	// Starting with Version 3.  Number of TIME_UNITS offset in first time point
    const timeOffset uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, timeOffset); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time offset %d, aborting: %v\n", timeOffset, err)
		os.Exit(1)
	}

	// Constant time step in TIME_UNITS. If set to zero, indicates variable time step.
	// 10,000 with TIME_UNITS of 0 would be 10 seconds.
	const zeroTimeStep uint32 = 0

	if err := binary.Write(fo, binary.LittleEndian, zeroTimeStep); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time step %d, aborting: %v\n", zeroTimeStep, err)
		os.Exit(1)
	}

	// Number of time points
    fmt.Fprintf(os.Stderr, "Number of time points: %d\n", len(times))
    //printTimes(field)
    timePoints := uint32(len(times))

	if err := binary.Write(fo, binary.LittleEndian, timePoints); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write time points count %d, aborting: %v\n", timePoints, err)
		os.Exit(1)
	}

    for i, t := range times {
		if err := binary.Write(fo, binary.LittleEndian, uint32(t - times[0])); err != nil {
            fmt.Fprintf(os.Stderr, "Unable to write time point %d, aborting: %v\n", i + 1, err)
			os.Exit(1)
		}
	}

 }

func main() {
	// maxSteps := flag.Int("times", 0, "Maximum number of time steps")
	// timeStep := flag.Float64("step", 60.0, "Time step in seconds")
	message := flag.String("message", "", "Set an optional message")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("%v", args)
		fmt.Fprintln(os.Stderr, "Usage: csv2b3d csvfile [b3dfile]")
		os.Exit(1)
	}

	csvFile := args[0]
    n := len(csvFile)
    b3dFile := csvFile[0:n-4] + ".b3d"

    if len(args) == 2 {
	    b3dFile = args[1]
    }

	fmt.Fprintf(os.Stderr, "Read: %s\n", b3dFile)

	fo, err := os.Create(b3dFile)
	defer fo.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %s for writing, aborting\n", b3dFile)
		os.Exit(1)
	}

	field := readFile(csvFile)
	fmt.Fprintf(os.Stderr, "Number of field points: %d\n", len(field))
    // printField(field)

	writeHeader(fo, field, *message)

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

	fmt.Fprintf(os.Stderr, "Wrote: %s\n", b3dFile)
}
