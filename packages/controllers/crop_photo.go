package controllers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/c-darwin/dcoin-go/packages/utils"
	"image"
	"image/jpeg"
	"os"
	"strings"
)

func (c *Controller) CropPhoto() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	photo := strings.Split(c.r.FormValue("photo"), ",")
	if len(photo) != 2 {
		return "", errors.New("Incorrect photo")
	}
	binary, err := base64.StdEncoding.DecodeString(photo[1])
	if err != nil {
		return "", err
	}
	img, _, err := image.Decode(bytes.NewReader(binary))
	if err != nil {
		return "", err
	}
	path := ""
	if c.r.FormValue("type") == "face" {
		path = *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg"
	} else {
		path = *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg"
	}
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	err = jpeg.Encode(out, img, &jpeg.Options{85})
	if err != nil {
		return "", err
	}

	return `{"success":"ok"}`, nil
}
