package detector

import (
	"encoding/json"
	"log"
)


func Race(path string) (string, error) {
	base := GetURL(DETECT)

	req, err := POSTtRequest(base.String(), path)
	if err != nil {
		return "", err
	}

	res, err := Send(req)
	if err != nil {
		return "", err
	}

	var data Data
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return "", err
	}

	return data.Face[0].Attr.Race.Value, err
}

func DetectSimilarity(faceId1, faceId2 string) (float64, error) {
	base := GetURL(COMPARE)
	params := base.Query()
	params.Add("face_id1", faceId1)
	params.Add("face_id2", faceId2)
	base.RawQuery = params.Encode()
	log.Println(base.String())
	req, err := GETRequest(base.String())
	if err != nil {
		return 0.0, err
	}

	res, err := Send(req)
	if err != nil {
		return 0.0, err
	}

	var sim Similarity
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&sim); err != nil {
		return 0.0, err
	}
	return sim.Similarity, err
}