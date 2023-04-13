package csv

import (
	"context"
	"encoding/json"
	"encore.dev/storage/sqldb"
)

func storeCountryData(ctx context.Context, csvRows []*CSVRow) error {
	tx, err := sqldb.Begin(ctx)
	if err != nil {
		return err
	}

	countryMap := map[string][]*CSVRow{}
	for _, row := range csvRows {
		countryMap[row.Country] = append(countryMap[row.Country], row)
	}

	var countryNameList []string
	for countryName := range countryMap {
		countryNameList = append(countryNameList, countryName)
	}
	countryNameListJSON, _ := json.Marshal(countryNameList)
	_, err = sqldb.Exec(
		ctx,
		`INSERT INTO "available_countries" (id, json) VALUES ($1, $2)
					ON CONFLICT (id) DO UPDATE SET json=$2`,
		"same_id_always", string(countryNameListJSON),
	)
	if err != nil {
		sqldb.Rollback(tx)
		return err
	}

	for countryName, countryRows := range countryMap {

		regionMap := map[string][]*CSVRow{}
		for _, row := range countryRows {
			regionMap[row.Region] = append(regionMap[row.Region], row)
		}

		var regions []*Region
		for regionName, regionRows := range regionMap {

			dayMap := map[string][]*CSVRow{}
			for _, row := range regionRows {
				dayMap[row.Day] = append(dayMap[row.Day], row)
			}

			var days []*Day
			for dayKey, csvRows := range dayMap {
				var times []*Time
				for _, csvRow := range csvRows {
					times = append(times, &Time{
						Hour:   csvRow.Hour,
						Offset: csvRow.Offset,
						Price:  csvRow.Price,
						Valid:  csvRow.Valid,
					})
				}
				days = append(days, &Day{
					Time: times,
					Date: dayKey,
				})
			}
			regions = append(regions, &Region{
				Name: regionName,
				Days: days,
			})
		}

		json, _ := json.Marshal(&Country{
			Name:    countryName,
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
	return nil
}
