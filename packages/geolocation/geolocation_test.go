package geolocation

import (
	"testing"
	"fmt"
)

func TestGetLocation(t *testing.T) {
	if _, err := GetLocation(); err != nil {
		fmt.Println(err.Error())
		t.Fatal(err)
	}
}
