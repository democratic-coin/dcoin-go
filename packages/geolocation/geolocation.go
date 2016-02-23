package geolocation

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"github.com/c-darwin/dcoin-go/packages/consts"
	//"runtime"
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
	var err error
	/*if runtime.GOOS == "darwin" {
		if coord, err := CLLocation(); err == nil {
			return coord, nil
		}

	} else
	if coord, err := getLocation(); err == nil && runtime.GOOS != "darwin" {
		return coord, nil
	}
*/
	return nil, err
}


func getLocation() (*coordinates, error) {
	var buf bytes.Buffer
	resp, err := http.Post("https://www.googleapis.com/geolocation/v1/geolocate?key=" + consts.GOOGLE_API_KEY, "json", &buf)
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
