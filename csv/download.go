package csv

import (
	"io"
	"net/http"
	"os"
)

func downloadFromURL(url string, fileName string) error {
	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			//return err
		}
	}(resp.Body)

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	return nil
}
