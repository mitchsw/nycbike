package importer

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Different dumps use different time formats.
const timeFormat1 = "2006-01-02 15:04:05"
const timeFormat2 = "1/2/2006 15:04:05"

func parseTime(timeStr string) (time.Time, error) {
	t, err := time.Parse(timeFormat1, timeStr)
	if err == nil {
		return t, nil
	}
	return time.Parse(timeFormat2, timeStr)
}

// A Trip is a parsed row from the Citi Bike System Data records.
type Trip struct {
	StartTime        time.Time
	StopTime         time.Time
	StartStationId   int
	StartStationName string
	StartStationLat  float64
	StartStationLong float64
	EndStationId     int
	EndStationName   string
	EndStationLat    float64
	EndStationLong   float64
}

// A TripdataReader downloaded, decompresses, and parses a CitiBike System Data file.
type TripdataReader struct {
	headerParsed        bool
	startTimeIdx        int
	stopTimeIdx         int
	startStationIdIdx   int
	startStationNameIdx int
	startStationLatIdx  int
	startStationLongIdx int
	endStationIdIdx     int
	endStationNameIdx   int
	endStationLatIdx    int
	endStationLongIdx   int

	files []io.ReadCloser
	csv   *csv.Reader
}

// Creates a new TripdataReader. This function downloads the file and decompresses it before returning.
func NewTripdataReader(zipUrl string) (*TripdataReader, error) {
	// Download and open the zip file.
	resp, err := http.Get(zipUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}

	// Open the csv files in the archive.
	r := &TripdataReader{
		headerParsed: false,
	}
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "_") {
			continue // Ignore weird __MACOX files in archives.
		}
		if !strings.HasSuffix(f.Name, ".csv") {
			continue
		}
		frc, err := f.Open()
		if err != nil {
			return nil, err
		}
		log.Printf("[tripdata_reader] Opened file: %v\n", f.Name)
		r.files = append(r.files, frc)
	}
	if len(r.files) == 0 {
		return nil, errors.New("expected .csv files in archive, found none")
	}

	// Setup to read the first file.
	r.csv = csv.NewReader(r.files[0])

	return r, nil
}

func (r *TripdataReader) Close() error {
	for _, f := range r.files {
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (r *TripdataReader) Read() (*Trip, error) {
	for {
		record, err := r.readRecord()
		if err != nil {
			return nil, err
		}
		trip, err := r.parseRecord(record)
		if err != nil {
			log.Printf("[tripdata_reader] Skipping trip: %v", err)
			continue
		}
		return trip, nil
	}
}

func (r *TripdataReader) parseHeader() error {
	record, err := r.csv.Read()
	if err != nil {
		return err
	}
	matches := 0
	for idx, col := range record {
		switch strings.ToLower(col) {
		case "starttime":
			fallthrough
		case "start time":
			r.startTimeIdx = idx
		case "stoptime":
			fallthrough
		case "stop time":
			r.stopTimeIdx = idx
		case "start station id":
			r.startStationIdIdx = idx
		case "start station name":
			r.startStationNameIdx = idx
		case "start station latitude":
			r.startStationLatIdx = idx
		case "start station longitude":
			r.startStationLongIdx = idx
		case "end station id":
			r.endStationIdIdx = idx
		case "end station name":
			r.endStationNameIdx = idx
		case "end station latitude":
			r.endStationLatIdx = idx
		case "end station longitude":
			r.endStationLongIdx = idx
		default:
			continue
		}
		matches++
	}
	if matches != 10 {
		return fmt.Errorf("expected 10 header matches, got %v: %v", matches, record)
	}
	r.headerParsed = true
	return nil
}

func (r *TripdataReader) readRecord() ([]string, error) {
	if !r.headerParsed {
		if err := r.parseHeader(); err != nil {
			return nil, err
		}
	}
	record, err := r.csv.Read()
	if err == io.EOF && len(r.files) > 1 {
		// Close file, move on to next file, and try again.
		if err := r.files[0].Close(); err != nil {
			return nil, err
		}
		r.files = r.files[1:]
		r.csv = csv.NewReader(r.files[0])
		r.headerParsed = false
		return r.readRecord()
	}
	return record, err
}

func (r *TripdataReader) parseRecord(record []string) (*Trip, error) {
	t := &Trip{}
	var err error
	t.StartTime, err = parseTime(record[r.startTimeIdx])
	if err != nil {
		return nil, fmt.Errorf("%w for StartTime; record: %+v", err, record)
	}
	t.StopTime, err = parseTime(record[r.stopTimeIdx])
	if err != nil {
		return nil, fmt.Errorf("%w for StopTime; record: %+v", err, record)
	}
	t.StartStationId, err = strconv.Atoi(record[r.startStationIdIdx])
	if err != nil {
		return nil, fmt.Errorf("%w for StartStationId; record: %+v", err, record)
	}
	t.StartStationName = record[r.startStationNameIdx]
	t.StartStationLat, err = strconv.ParseFloat(record[r.startStationLatIdx], 64)
	if err != nil {
		return nil, fmt.Errorf("%w for StartStationName; record: %+v", err, record)
	}
	t.StartStationLong, err = strconv.ParseFloat(record[r.startStationLongIdx], 64)
	if err != nil {
		return nil, fmt.Errorf("%w for StartStationLong; record: %+v", err, record)
	}
	t.EndStationId, err = strconv.Atoi(record[r.endStationIdIdx])
	if err != nil {
		return nil, fmt.Errorf("%w for EndStationId; record: %+v", err, record)
	}
	t.EndStationName = record[r.endStationNameIdx]
	t.EndStationLat, err = strconv.ParseFloat(record[r.endStationLatIdx], 64)
	if err != nil {
		return nil, fmt.Errorf("%w for EndStationLat; record: %+v", err, record)
	}
	t.EndStationLong, err = strconv.ParseFloat(record[r.endStationLongIdx], 64)
	if err != nil {
		return nil, fmt.Errorf("%w for EndStationLong; record: %+v", err, record)
	}
	return t, nil
}
