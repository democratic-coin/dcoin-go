package geolocation


type Location struct {
	Coordinates *coordinates `json:"location"`
	Accuracy    float32 `json:"accuracy"`
}

type coordinates struct {
	Latitude float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

func GetLocation() (*coordinates, error) {
	coord, err := getLocation()
	return coord, err
}
