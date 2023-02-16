package forecast

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"net/http"
	"strings"
)

//go:embed forecast.xlsx
var excelFile embed.FS

//go:embed forecast.json
var jsonFile embed.FS

func readExcelFile() {
	str, err := excelFile.ReadFile("forecast.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := excelize.OpenReader(strings.NewReader(string(str)))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Get value from cell by given worksheet name and cell reference.
	cell, err := f.GetCellValue("Sheet1", "B2")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cell)
	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}

type DataPoint struct {
	Year  string `json:"year"`
	Price string `json:"price"`
}

type ZoneForecast struct {
	Zone string       `json:"zone"`
	Name string       `json:"name"`
	Data []*DataPoint `json:"data"`
}

type ListResponse struct {
	Zones []*ZoneForecast `json:"zones"`
}

func readJsonFile() ([]*ZoneForecast, error) {
	str, err := jsonFile.ReadFile("forecast.json")
	if err != nil {
		return nil, err
	}
	var forecastData []*ZoneForecast
	json.Unmarshal(str, &forecastData)

	return forecastData, nil
}

//encore:api public path=/forecasts
func GetForecasts(ctx context.Context) (*ListResponse, error) {
	forecastData, err := readJsonFile()
	if err != nil {
		return nil, err
	}
	return &ListResponse{
		Zones: forecastData,
	}, nil
}

type PostalCodeInfo struct {
	Status  int    `json:"status"`
	Message string `json:"msg"`
	Zone    string `json:"zone"`
}

func lookupPostalCode(postalCode string) (*PostalCodeInfo, error) {
	URL := "https://elpriset.nu/wp-json/zones/v1/zone/" + postalCode
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response *PostalCodeInfo

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	fmt.Println(response)

	return response, nil
}

//encore:api public path=/zone/:postalCode
func GetZoneFromPostalCode(ctx context.Context, postalCode string) (*PostalCodeInfo, error) {
	info, _ := lookupPostalCode(postalCode)
	return &PostalCodeInfo{
		Status:  info.Status,
		Message: info.Message,
		Zone:    info.Zone,
	}, nil
}
