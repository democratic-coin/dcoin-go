package detector

import (
	"net/url"
	"log"
	"net/http"
	"bytes"
	"mime/multipart"
	"os"
	"io"
	"encoding/json"
)

const (
	BASE_URL = "http://apius.faceplusplus.com/v2/detection/detect"
	API_KEY = "6ee56f855de7aaf3890bc2a20e006b7a"
	API_SECRET = "xvieJyM1i_aQ4J1oudxcsCdHenviBI_P"
)


func formRequest(url, file string) (*http.Request, error) {
	var buf bytes.Buffer
	nWriter := multipart.NewWriter(&buf)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	writer, err := nWriter.CreateFormFile("img", file)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(writer, f); err != nil {
		return nil, err
	}
	if writer, err = nWriter.CreateFormField("img"); err != nil {
		return nil, err
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	nWriter.Close()

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	//Set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", nWriter.FormDataContentType())
	return req, err
}

func request(url, file string) (string, error) {

	req, err := formRequest(url, file)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		log.Fatalf("bad status: %s\n", res.Status)
	}

	var data Data
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return "", err
	}

	return data.Face[0].Attr.Race.Value, err
}

func Detect(path string) (string, error) {
	base, _ := url.Parse(BASE_URL)
	params := url.Values{}
	params.Add("api_key", API_KEY)
	params.Add("api_secret", API_SECRET)
	base.RawQuery = params.Encode()

	race, err := request(base.String(), path)
	if err != nil {
		return "", err
	}

	return race, err
}