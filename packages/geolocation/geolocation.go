package geolocation

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
)

type Location struct {
	Coordinates *coordinates `json:"location"`
	Accuracy    float32 `json:"accuracy"`
}

type coordinates struct {
	Latitude float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

func GetLocation() (*coordinates, error) {
	var buf bytes.Buffer
	resp, err := http.Post("https://www.googleapis.com/geolocation/v1/geolocate?key=AIzaSyBLZlUPgd9uhX05OrsFU68yJOZFrYhZe84", "json", &buf)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Cannot read body:", err.Error())
	}

	loc, err := parseResponse(body)
	if err != nil {
		fmt.Println("Cannot parse:", err.Error())
		return nil, err
	}
	return loc.Coordinates, nil
}

func parseResponse(b []byte) (*Location, error) {
	var pos *Location
	if err := json.Unmarshal(b, &pos); err != nil {
		return nil, err
	}

	return pos, nil
}
