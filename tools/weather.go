// Package tools
// Author: Evsikov Artem

package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const WeatherAPIKey = "61a57e3c23b124626eb864b04fb6e27f"

type WeatherData struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	// Otras propiedades relevantes para el clima
}

type WeatherDataResponce struct {
	City string  `json:"city"`
	Temp float64 `json:"temp"`
}

type WeatherParams struct {
	City string `json:"city"`
}

type Location struct {
	Name       string            `json:"name"`
	LocalNames map[string]string `json:"local_names"`
	Lat        float64           `json:"lat"`
	Lon        float64           `json:"lon"`
	Country    string            `json:"country"`
	State      string            `json:"state"`
}

func RequestLocation(query string) (float64, float64) {
	url := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", query, WeatherAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return 0, 0
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the response:", err)
		return 0, 0
	}

	var locations []Location
	json.Unmarshal(body, &locations)

	if len(locations) > 0 {
		return locations[0].Lat, locations[0].Lon
	}
	return 0, 0
}

func RequestWeather(arguments string) (string, error) {
	var location map[string]string
	json.NewDecoder(bytes.NewReader([]byte(arguments))).Decode(&location)
	lat, lon := RequestLocation(location["location"])
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric",
		lat, lon, WeatherAPIKey)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return "Error while requesting weather", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the response:", err)
		return "Error while requesting weather", err
	}

	return string(body), nil
}
