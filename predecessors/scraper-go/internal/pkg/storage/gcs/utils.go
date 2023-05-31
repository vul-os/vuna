package storage

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

func WriteTextFile(writer io.Writer, data []string) error {
	for _, line := range data {
		_, err := fmt.Fprintln(writer, line)
		if err != nil {
			return fmt.Errorf("error writing data to GCS: %v", err)
		}
	}

	return nil
}

func WriteCSVFile(writer io.Writer, data []map[string]string) error {
	csvWriter := csv.NewWriter(writer)

	// Write CSV header
	if len(data) == 0 {
		return nil // No data to write
	}

	header := make([]string, 0, len(data[0]))
	for key := range data[0] {
		header = append(header, key)
	}

	// Sort the header keys
	sort.Strings(header)
	csvWriter.Write(header)

	// Write CSV data rows
	for _, row := range data {
		record := make([]string, 0, len(header))
		for _, key := range header {
			fmt.Println(row[key])
			record = append(record, row[key])
		}
		csvWriter.Write(record)
	}

	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("error writing data to CSV: %v", err)
	}

	return nil
}

func ReadTextFile(reader io.Reader) ([]string, error) {
	lines := make([]string, 0)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading text file: %v", err)
	}

	return lines, nil
}

func ReadCSVFile(reader io.Reader) ([]map[string]string, error) {
	csvReader := csv.NewReader(reader)
	rows := make([]map[string]string, 0)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV headers: %v", err)
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error reading CSV record: %v", err)
		}

		row := make(map[string]string)
		for i, value := range record {
			row[headers[i]] = value
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func GetFileExtension(filePath string) string {
	fileName := filepath.Base(filePath)
	extension := filepath.Ext(fileName)
	return strings.ToLower(extension)
}
