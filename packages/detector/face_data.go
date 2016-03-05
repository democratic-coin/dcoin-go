package detector


type Data struct {
	Face []Face

	ImgHeight int `json:"img_height"`
	ImgID     string `json:"img_id"`
	ImgWidth  int `json:"img_width"`
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

type Face struct {
	FaceID    string `json:"face_id"`
	Attr 	Attribute `json:"attribute"`
	Pos 	Position `json:"position"`
	Tag string
}

type Attribute struct {
	Age struct {
		Range int
		Value int
	} `json:"age"`
	Gender struct {
		Confidence float64
		Value      string
	} `json:"gender"`
	Race struct {
		Confidence float64
		Value      string
	} `json:"race"`
	Smiling struct {
		Value float64
	}
}

type Position struct {
	Center struct {
		X, Y float64
	}
	EyeLeft struct {X, Y float64 } `json:"eye_left"`
	EyeRight struct {X, Y float64 } `json:"eye_right"`
	Height    float64
	MouthLeft struct {X, Y float64 } `json:"mouth_left"`
	MouthRight struct {X, Y float64 } `json:"mouth_right"`
	Nose struct {
		X, Y float64
	}
	Width float64
}