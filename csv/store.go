package csv

import (
	"context"
	"encoding/json"
	"encore.dev/storage/sqldb"
	"errors"
	"time"
)

var countryNameToIsoMap = map[string]string{
	"MONTENEGRO":             "ME",
	"NETHERLANDS":            "NL",
	"CYPRUS":                 "CY",
	"BELGIUM":                "BE",
	"UNITED KINGDOM":         "GB",
	"BELARUS":                "BY",
	"GERMANY":                "DE",
	"NORTH MACEDONIA":        "MK",
	"POLAND":                 "PL",
	"SLOVENIA":               "SI",
	"FRANCE":                 "FR",
	"ITALY":                  "IT",
	"CROATIA":                "HR",
	"HUNGARY":                "HU",
	"IRELAND":                "IE",
	"PORTUGAL":               "PT",
	"SERBIA":                 "RS",
	"RUSSIA":                 "RU",
	"CZECH REPUBLIC":         "CZ",
	"GREECE":                 "GR",
	"MALTA":                  "MT",
	"LUXEMBOURG":             "LU",
	"SLOVAKIA":               "SK",
	"LATVIA":                 "LV",
	"ALBANIA":                "AL",
	"ICELAND":                "IS",
	"MOROCCO":                "MA",
	"DENMARK":                "DK",
	"FINLAND":                "FI",
	"SPAIN":                  "ES",
	"ROMANIA":                "RO",
	"TURKEY":                 "TR",
	"MOLDOVA":                "MD",
	"BULGARIA":               "BG",
	"SWITZERLAND":            "CH",
	"ESTONIA":                "EE",
	"LITHUANIA":              "LT",
	"NORWAY":                 "NO",
	"SWEDEN":                 "SE",
	"AUSTRIA":                "AT",
	"BOSNIA AND HERZEGOVINA": "BA",
	"KOSOVO":                 "XK",
}

func containsValidData(rows []*CSVRow) bool {
	for _, row := range rows {
		if row.Valid != "0" {
			return true
		}
	}
	return false
}

func storeCountryData(ctx context.Context, csvRows []*CSVRow) error {
	tx, err := sqldb.Begin(ctx)
	if err != nil {
		return err
	}

	timezone, _ := time.LoadLocation("Europe/Stockholm")

	countryMap := map[string][]*CSVRow{}
	for _, row := range csvRows {
		countryMap[row.Country] = append(countryMap[row.Country], row)
	}

	var countrySimpleList []*CountrySimple

	for countryName, countryRows := range countryMap {
		hasValidData := containsValidData(countryRows)
		if !hasValidData {
			continue
		}

		regionMap := map[string][]*CSVRow{}
		for _, row := range countryRows {
			regionMap[row.Region] = append(regionMap[row.Region], row)
		}

		var regions []*Region
		var regionNames []string
		for regionName, regionRows := range regionMap {
			hasValidData = containsValidData(regionRows)
			if !hasValidData {
				println(countryName, regionName)
				continue
			}
			regionNames = append(regionNames, regionName)

			dayMap := map[string][]*CSVRow{}
			for _, row := range regionRows {
				dayMap[row.LocalDay] = append(dayMap[row.LocalDay], row)
			}

			var days []*Day
			for dayKey, csvRows := range dayMap {
				var times []*Time

				// Skip days that have fewer than 10 hours
				if len(csvRows) > 10 {
					for _, csvRow := range csvRows {
						times = append(times, &Time{
							Hour:       csvRow.Hour,
							LocalHour:  csvRow.LocalHour,
							Offset:     csvRow.Offset,
							Price:      csvRow.Price,
							LocalPrice: csvRow.LocalPrice,
							Valid:      csvRow.Valid,
						})
					}
					days = append(days, &Day{
						Time: times,
						Date: dayKey,
					})
				}
			}
			var daysAfterToday []*Day
			for _, day := range days {
				// Format the date to be able to parse it
				dateFormatted := day.Date[0:4] + "-" + day.Date[4:6] + "-" + day.Date[len(day.Date)-2:]
				date, _ := time.Parse("2006-01-02", dateFormatted)
				// If the date is after today in the given timezone, add it to the list
				now := time.Now().In(timezone)
				isToday := now.Year() == date.Year() && now.Month() == date.Month() && now.Day() == date.Day()
				if (isToday || date.After(now)) && len(daysAfterToday) <= 7 {
					daysAfterToday = append(daysAfterToday, day)
				}
			}
			regions = append(regions, &Region{
				Name:     regionName,
				Currency: regionRows[0].Currency,
				Days:     daysAfterToday,
			})
		}

		isoCode, ok := countryNameToIsoMap[countryName]
		if !ok {
			return errors.New("Country " + countryName + " does not exist in map")
		}

		countrySimpleList = append(countrySimpleList, &CountrySimple{
			Name:    countryName,
			ISOCode: isoCode,
			Regions: regionNames,
		})

		json, _ := json.Marshal(&Country{
			Name:    countryName,
			ISOCode: isoCode,
			Regions: regions,
		})

		_, err := sqldb.Exec(
			ctx,
			`INSERT INTO "country_data" (countryName, json) VALUES ($1, $2)
			ON CONFLICT (countryName) DO UPDATE SET json=$2`,
			countryName, string(json),
		)
		if err != nil {
			sqldb.Rollback(tx)
			return err
		}
	}

	countrySimpleListJSON, _ := json.Marshal(countrySimpleList)
	_, err = sqldb.Exec(
		ctx,
		`INSERT INTO "available_countries" (id, json) VALUES ($1, $2)
					ON CONFLICT (id) DO UPDATE SET json=$2`,
		"same_id_always", string(countrySimpleListJSON),
	)
	if err != nil {
		sqldb.Rollback(tx)
		return err
	}

	return nil
}
