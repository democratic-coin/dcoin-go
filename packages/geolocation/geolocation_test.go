package geolocation

import (
	"testing"
)

func TestGetLocation(t *testing.T) {
	if _, err := GetLocation(); err != nil {
		t.Fatal(err)
	}
}