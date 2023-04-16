package csv

import (
	"context"
	"encoding/json"
	"encore.app/slack"
	"encore.dev/cron"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/getsentry/sentry-go"
	"os"
	"path/filepath"
	"time"
)

type Time struct {
	Hour   string `json:"hour"`
	Offset string `json:"offset"`
	Price  string `json:"price"`
	Valid  string `json:"valid"`
}

type Day struct {
	Date string  `json:"date"`
	Time []*Time `json:"time"`
}

type Region struct {
	Name string `json:"name"`
	Days []*Day `json:"days"`
}

type Country struct {
	Name    string    `json:"name"`
	ISOCode string    `json:"isoCode"`
	Regions []*Region `json:"regions"`
}

type CountrySimple struct {
	Name    string   `json:"name"`
	ISOCode string   `json:"isoCode"`
	Regions []string `json:"regions"`
}

var DROPBOX_URL = "https://dropbox.com/sh/0qkvdyeychvde9o/AACw-J_gPfhpPF6kFMcFjK-xa?dl=1"
var ZIP_FILE_NAME = "file.zip"
var ZIP_FOLDER = "data"
var CSV_FILE_NAME = "EVIEW_Price.csv"

var _ = cron.NewJob("download-csv-file", cron.JobConfig{
	Title:    "Download a new CSV file from Dropbox",
	Endpoint: SaveCsv,
	Every:    1 * cron.Hour,
})

//encore:api public method=GET path=/csv
func SaveCsv(ctx context.Context) error {
	pwd, err := os.Getwd()
	if err != nil {
		captureError(err)
		return err
	}

	err = downloadFromURL(DROPBOX_URL, ZIP_FILE_NAME)
	if err != nil {
		captureError(err)
		return err
	}

	err = unzipFile(ZIP_FILE_NAME, filepath.Join(pwd, ZIP_FOLDER))
	if err != nil {
		captureError(err)
		return err
	}

	csvRows, err := parseCSVFile(filepath.Join(ZIP_FOLDER, CSV_FILE_NAME))
	if err != nil {
		captureError(err)
		return err
	}

	err = storeCountryData(ctx, csvRows)
	if err != nil {
		captureError(err)
		return err
	}

	return nil
}

type StoredCountry struct {
	Name string
	JSON string
}

type GetCountryResponse struct {
	Data *Country `json:"data"`
}

//encore:api public method=GET path=/country/:countryName
func GetCountry(ctx context.Context, countryName string) (*GetCountryResponse, error) {
	stored := &StoredCountry{}

	err := sqldb.QueryRow(
		ctx, `SELECT countryName, json FROM "country_data" WHERE countryName=$1`, countryName,
	).Scan(&stored.Name, &stored.JSON)
	if err != nil {
		captureError(err)
		return nil, err
	}

	var country Country
	if err = json.Unmarshal([]byte(stored.JSON), &country); err != nil {
		captureError(err)
		return nil, err
	}

	return &GetCountryResponse{Data: &country}, nil
}

type StoredCountryList struct {
	ID   string
	JSON string
}

type GetCountryListResponse struct {
	Data *[]CountrySimple `json:"data"`
}

//encore:api public method=GET path=/country
func GetCountryList(ctx context.Context) (*GetCountryListResponse, error) {
	stored := &StoredCountryList{}

	err := sqldb.QueryRow(
		ctx, `SELECT id, json FROM "available_countries" WHERE id=$1`, "same_id_always",
	).Scan(&stored.ID, &stored.JSON)
	if err != nil {
		captureError(err)
		return nil, err
	}

	var countrySimpleList *[]CountrySimple
	if err = json.Unmarshal([]byte(stored.JSON), &countrySimpleList); err != nil {
		captureError(err)
		return nil, err
	}

	return &GetCountryListResponse{Data: countrySimpleList}, nil
}

func sendToSlack(ctx context.Context, message string) {
	err := slack.Notify(ctx, &slack.NotifyParams{Text: message})
	if err != nil {
		rlog.Error(err.Error())
	}
}

func initSentry() {
	if sentry.CurrentHub().Client() == nil {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: "https://ca180dbb35ae42dfad0515b04c921b71@o4505000914452480.ingest.sentry.io/4505000915697664",
			// Set TracesSampleRate to 1.0 to capture 100%
			// of transactions for performance monitoring.
			// We recommend adjusting this value in production,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			rlog.Error("sentry.Init: %s", err)
		}
	}
}

func captureMessage(msg string) {
	initSentry()
	defer sentry.Flush(2 * time.Second)
	sentry.CaptureMessage(msg)
}

func captureError(err error) {
	initSentry()
	defer sentry.Flush(2 * time.Second)
	sentry.CaptureException(err)
}
