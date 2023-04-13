package forecast

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

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

type PostalCodeInfo struct {
	Status  int    `json:"status"`
	Message string `json:"msg"`
	Zone    string `json:"zone"`
}

func lookupPostalCode(postalCode string) (*PostalCodeInfo, error) {
	URL := "https://elpriset.nu/wp-json/zones/v1/zone/" + url.PathEscape(postalCode)
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response *PostalCodeInfo

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
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

type IPLocationInfo struct {
	ZipCode string `json:"zip_code"`
}

func lookupIp(ip string) (*IPLocationInfo, error) {
	URL := "https://api.ip2location.com/v2/?ip=" + ip + "&package=WS9&format=json&key=RF0AZKKB5I"
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response *IPLocationInfo

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

type Request struct {
	Ip string `json:"ip"`
}

//encore:api public method=POST path=/ip
func GetPostalCodeFromIP(ctx context.Context, req *Request) (*IPLocationInfo, error) {
	info, _ := lookupIp(req.Ip)
	return &IPLocationInfo{
		ZipCode: info.ZipCode,
	}, nil
}
