package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type Kind int

const (
	bold = "\x1b[1m"

	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	gray    = "\x1b[90m"
	reset   = "\x1b[0m"
)
const (
	Sunny Kind = iota
	Cloudy
	MostlyCloudy
	PartlyCloudy
	Rainy
	LightRain
	Snowy
	Foggy
	Lightning
	Hail
)

func (k Kind) String() string {
	switch k {
	case Sunny:
		return "sunny"
	case Cloudy:
		return "cloudy"
	case MostlyCloudy:
		return "mostly cloudy"
	case PartlyCloudy:
		return "party cloudy"
	case Rainy:
		return "rainy"
	case LightRain:
		return "light rain"
	case Snowy:
		return "snowy"
	case Foggy:
		return "foggy"
	case Lightning:
		return "lightning"
	case Hail:
		return "hail"
	default:
		return "cloudy"
	}
}

type rules struct {
	lightning []*regexp.Regexp
	rain      []*regexp.Regexp
	snow      []*regexp.Regexp
	wintry    []*regexp.Regexp
	hail      []*regexp.Regexp
	fog       []*regexp.Regexp
	wind      []*regexp.Regexp
	sunny     []*regexp.Regexp
	cloudy    []*regexp.Regexp
}

func mustAll(patterns ...string) []*regexp.Regexp {
	var regexes []*regexp.Regexp
	for _, p := range patterns {
		regexes = append(regexes, regexp.MustCompile(p))
	}
	return regexes
}

var rx = rules{
	// Thunder / lightning
	lightning: mustAll(
		`(?i)\b(thunderstorm|thunderstorms|thunder)\b`,
		`(?i)\b(t[\s-]?storms?|t[\s-]?storm)\b`,
		`(?i)\b(tsra|vcts)\b`,
		`(?i)\b(squall)\b`,
		`(?i)\b(lightning)\b`,
	),

	// Rain-ish (includes drizzle/showers)
	rain: mustAll(
		`(?i)\b(rain|rains|rainfall)\b`,
		`(?i)\b(showers?|rain\s*showers?)\b`,
		`(?i)\b(drizzle|sprinkles?)\b`,
		`(?i)\b(downpour|pouring)\b`,
		`(?i)\b(precip(itation)?)\b`,
		`(?i)\b(chance\s+rain|likely\s+rain|periods?\s+of\s+rain)\b`,
		`(?i)\b(scattered\s+showers?|isolated\s+showers?)\b`,
		`(?i)\b(slight\s+chance\s+showers?)\b`,
	),

	// Snow-ish
	snow: mustAll(
		`(?i)\b(snow|snows|snowfall)\b`,
		`(?i)\b(flurries|snow\s*flurries)\b`,
		`(?i)\b(snow\s*showers?)\b`,
		`(?i)\b(blizzard)\b`,
		`(?i)\b(blowing\s+snow|drifting\s+snow)\b`,
		`(?i)\b(snow\s*squalls?)\b`,
		`(?i)\b(lake[-\s]?effect\s+snow)\b`,
	),

	// Wintry mix / ice types (often not "snowy" or "rainy" cleanly)
	wintry: mustAll(
		`(?i)\b(wintry\s+mix)\b`,
		`(?i)\b(rain\s*/\s*snow|snow\s*/\s*rain)\b`,
		`(?i)\b(mix(ed)?\s+precip(itation)?)\b`,
		`(?i)\b(sleet)\b`,
		`(?i)\b(ice\s+pellets?)\b`,
		`(?i)\b(freezing\s+rain)\b`,
		`(?i)\b(freezing\s+drizzle)\b`,
		`(?i)\b(glaze|icing)\b`,
	),

	// Hail
	hail: mustAll(
		`(?i)\b(hail)\b`,
		`(?i)\b(small\s+hail)\b`,
		`(?i)\b(graupel)\b`,
	),

	// Fog / low vis
	fog: mustAll(
		`(?i)\b(fog|foggy)\b`,
		`(?i)\b(patchy\s+fog)\b`,
		`(?i)\b(dense\s+fog)\b`,
		`(?i)\b(mist|misty)\b`,
		`(?i)\b(low\s+visibility|reduced\s+visibility)\b`,
	),

	// Wind
	wind: mustAll(
		`(?i)\b(windy)\b`,
		`(?i)\b(breezy)\b`,
		`(?i)\b(gusty|gusts?)\b`,
		`(?i)\b(blustery)\b`,
		`(?i)\b(strong\s+winds?)\b`,
		`(?i)\b(high\s+winds?)\b`,
	),

	// Sunny / clear-ish
	sunny: mustAll(
		`(?i)\b(sunny)\b`,
		`(?i)\b(clear)\b`,
		`(?i)\b(fair)\b`,
		`(?i)\b(mostly\s+sunny)\b`,
		`(?i)\b(sunshine)\b`,
		`(?i)\b(becoming\s+sunny)\b`,
		`(?i)\b(sunny\s+and\s+warm)\b`,
	),

	// Cloudy (includes partly/mostly)
	cloudy: mustAll(
		`(?i)\b(cloudy)\b`,
		`(?i)\b(mostly\s+cloudy)\b`,
		`(?i)\b(partly\s+cloudy)\b`,
		`(?i)\b(increasing\s+clouds?)\b`,
		`(?i)\b(decreasing\s+clouds?)\b`,
		`(?i)\b(overcast)\b`,
		`(?i)\b(broken\s+clouds?)\b`,
		`(?i)\b(scattered\s+clouds?)\b`,
	),
}

func matchesAny(text string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func classifyWeather(description string) Kind {
	// Check in priority order (most specific first)
	if matchesAny(description, rx.lightning) {
		return Lightning
	}

	if matchesAny(description, rx.hail) {
		return Hail
	}

	if matchesAny(description, rx.snow) {
		return Snowy
	}

	if matchesAny(description, rx.rain) {
		// Check if it's light rain specifically
		if matchesAny(description, mustAll(`(?i)\b(light|slight|drizzle|sprinkle)\b`)) {
			return LightRain
		}
		return Rainy
	}

	if matchesAny(description, rx.fog) {
		return Foggy
	}

	if matchesAny(description, rx.sunny) {
		return Sunny
	}

	if matchesAny(description, rx.cloudy) {
		// Check if mostly or partly cloudy
		if matchesAny(description, mustAll(`(?i)\b(mostly)\b`)) {
			return MostlyCloudy
		}
		if matchesAny(description, mustAll(`(?i)\b(partly)\b`)) {
			return PartlyCloudy
		}
		return Cloudy
	}

	return Cloudy
}

func changeTempColor(temp float32, tempUnit string) string {
	var colorCode string
	if tempUnit == "F" {
		switch {
		case temp <= 32:
			colorCode = blue
		case temp > 32 && temp <= 60:
			colorCode = cyan
		case temp > 60 && temp <= 80:
			colorCode = green
		case temp > 80 && temp <= 90:
			colorCode = yellow
		case temp > 90:
			colorCode = red
		}
	} else {
		switch {
		case temp <= 0:
			colorCode = blue
		case temp > 0 && temp <= 15:
			colorCode = cyan
		case temp > 15 && temp <= 27:
			colorCode = green
		case temp > 27 && temp <= 32:
			colorCode = yellow
		case temp > 32:
			colorCode = red
		}
	}
	return colorCode
}

func mapWeatherToColor(weatherType string) string {
	switch weatherType {
	case "sunny":
		return yellow
	case "cloudy", "mostly cloudy", "partly cloudy":
		return gray
	case "rainy", "light rain":
		return blue
	case "snowy":
		return cyan
	case "foggy":
		return gray
	case "lightning", "hail":
		return magenta
	default:
		return gray
	}
}

func displayArt(weatherType string, temp float32, tempUnit string, description string) {
	// Try multiple paths for the JSON file
	possiblePaths := []string{
		"designs/weather.json",
		"./designs/weather.json",
		filepath.Join(filepath.Dir(os.Args[0]), "designs", "weather.json"),
		filepath.Join(filepath.Dir(os.Args[0]), "..", "designs", "weather.json"),
	}

	var data []byte
	var err error

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		fmt.Printf("Weather: %s\n", description)
		fmt.Printf("Temperature: %.0f %s\n", temp, tempUnit)
		return
	}

	var weatherData map[string]map[string]string
	err = json.Unmarshal(data, &weatherData)
	if err != nil {
		fmt.Printf("Weather: %s\n", description)
		fmt.Printf("Temperature: %.0f %s\n", temp, tempUnit)
		return
	}

	weatherColor := mapWeatherToColor(weatherType)
	// Get the picture for the weather type
	if weatherInfo, ok := weatherData[weatherType]; ok {
		if picture, ok := weatherInfo["picture"]; ok {
			// Apply color to ASCII art - print color code, then art, then reset
			fmt.Print(weatherColor)
			fmt.Println(picture)
			fmt.Print(reset)
		}
	}

	tempColorCode := changeTempColor(temp, tempUnit)

	// Display weather data with hard-forced colors
	fmt.Printf("\n%s%s%s\n", tempColorCode, description, reset)
	fmt.Printf("Temperature: %s%.0f %s%s\n", tempColorCode, temp, tempUnit, reset)
}
