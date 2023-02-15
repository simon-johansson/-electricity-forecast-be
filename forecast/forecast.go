package forecast

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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

//encore:api public path=/forecast/:postalCode
func Forecast(ctx context.Context, postalCode string) (*PostalCodeInfo, error) {
	info, _ := lookupPostalCode(postalCode)
	return &PostalCodeInfo{
		Status:  info.Status,
		Message: info.Message,
		Zone:    info.Zone,
	}, nil
}
