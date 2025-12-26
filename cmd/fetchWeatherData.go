package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type LocationData struct {
	Results []struct {
		Id          int64   `json:"id"`
		Name        string  `json:"name"`
		Latitude    float32 `json:"latitude"`
		Longitude   float32 `json:"longitude"`
		Elevation   float32 `json:"elevation"`
		CountryCode string  `json:"country_code"`
		Timezone    string  `json:"timezone"`
		Population  int64   `json:"population"`
		Country     string  `json:"country"`
	} `json:"results"`
}

type pointsResp struct {
	Properties struct {
		Forecast string `json:"forecast"`
		GridId   string `json:"gridId"`
		GridX    int    `json:"gridX"`
		GridY    int    `json:"gridY"`
	} `json:"properties"`
}

type forecastResp struct {
	Properties struct {
		Periods []struct {
			Name            string `json:"name"`
			Temperature     int    `json:"temperature"`
			TemperatureUnit string `json:"temperatureUnit"`
			ShortForecast   string `json:"shortForecast"`
			WindSpeed       string `json:"windSpeed"`
			WindDirection   string `json:"windDirection"`
		} `json:"periods"`
	} `json:"properties"`
}

func getLatLong(city string) (lat float32, long float32) {

	client := &http.Client{Timeout: 10 * time.Second}

	// Properly URL encode the city name
	encodedCity := url.QueryEscape(city)
	req, err := client.Get("https://geocoding-api.open-meteo.com/v1/search?name=" + encodedCity)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	defer req.Body.Close()

	// Check HTTP status code
	if req.StatusCode != http.StatusOK {
		fmt.Printf("API returned status %d\n", req.StatusCode)
		return
	}

	var location LocationData

	err = json.NewDecoder(req.Body).Decode(&location)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	if len(location.Results) == 0 {
		fmt.Println("No results found for city:", city)
		return
	}

	var latitude = location.Results[0].Latitude
	var longitude = location.Results[0].Longitude

	// fmt.Println("Latitude:", latitude)
	// fmt.Println("Longitude:", longitude)

	return latitude, longitude
}

func fetchWeatherData(lat float32, long float32) (temp float32, tempUnit string, desc string) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := client.Get(fmt.Sprintf("https://api.weather.gov/points/%f,%f", lat, long))

	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	defer req.Body.Close()

	var points pointsResp
	err = json.NewDecoder(req.Body).Decode(&points)
	if err != nil {
		fmt.Println("Error decoding points response:", err)
		return
	}

	forecastReq, err := client.Get(points.Properties.Forecast)

	if err != nil {
		//fmt.Println("Error creating forecast request:", err)
		return
	}

	defer forecastReq.Body.Close()

	var forecast forecastResp

	err = json.NewDecoder(forecastReq.Body).Decode(&forecast)

	if err != nil {
		//fmt.Println("Error decoding forecast response:", err)
		return
	}

	for _, period := range forecast.Properties.Periods {
		//fmt.Printf("%s: %d %s, %s, Wind: %s %s\n", period.Name, period.Temperature, period.TemperatureUnit, period.ShortForecast, period.WindSpeed, period.WindDirection)
		return float32(period.Temperature), period.TemperatureUnit, period.ShortForecast
	}

	return 0, "", ""
}
