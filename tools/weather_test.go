package tools

import (
	"testing"
)

// TestGetWeather request weather
func TestGetWeather(t *testing.T) {
	_, err := RequestWeather("moscow")
	if err != nil {
        t.Fatalf("Error requesting weather: %s", err)
	}
}
