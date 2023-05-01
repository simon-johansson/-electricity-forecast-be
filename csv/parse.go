package csv

import (
	"encoding/csv"
	"encore.dev/rlog"
	"github.com/gocarina/gocsv"
	"io"
	"os"
)

type CSVRow struct {
	Time    string `csv:"CALCTIME"`
	Country string `csv:"COUNTRY"`
	Region  string `csv:"REGION"`
	Day     string `csv:"DAG"`
	Hour    string `csv:"TIMMA"`
	Offset  string `csv:"OFFSET"`
	Valid   string `csv:"GILTLIG"`
	Price   string `csv:"PRIS"`
}

func parseCSVFile(filePath string) ([]*CSVRow, error) {
	clientsFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()

	clients := []*CSVRow{}

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		//r.FieldsPerRecord = -1
		r.Comma = ';'
		return r
	})

	if err = gocsv.UnmarshalFile(clientsFile, &clients); err != nil {
		rlog.Error("Failed to unmarshal file", err)
		return nil, err
	}

	return clients, nil
}
