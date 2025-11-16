package logic

import (
	"math/rand"
	"time"
)

// OutdoorStats holds the data you want to expose to your system.
type OutdoorStats struct {
	TemperatureF    float64   `json:"temperature_f"`
	Humidity        float64   `json:"humidity"`         // relative humidity (%)
	PrecipitationMM float64   `json:"precipitation_mm"` // precipitation (mm)
	WindSpeedMph    float64   `json:"wind_speed_mph"`
	Timestamp       time.Time `json:"timestamp"`
}

// FetchOutdoorStats generates realistic random outdoor weather statistics.
func FetchOutdoorStats() (*OutdoorStats, error) {
	// Temperature: 20-95°F (realistic outdoor range)
	tempF := float64(rand.Intn(76) + 20)

	// Humidity: 20-90% (realistic outdoor humidity range)
	humidity := float64(rand.Intn(71) + 20)

	// Precipitation: Most days are dry (0mm), occasionally some precipitation (0-25mm)
	// 70% chance of no precipitation, 30% chance of some precipitation
	var precipMM float64
	if rand.Intn(100) < 70 {
		precipMM = 0.0
	} else {
		// When precipitation occurs, range from 0.1mm to 25mm
		precipMM = float64(rand.Intn(250)+1) / 10.0
	}

	// Wind speed: 0-25 mph (realistic wind speed range)
	windMph := float64(rand.Intn(26))

	return &OutdoorStats{
		TemperatureF:    roundToOneDecimal(tempF),
		Humidity:        roundToOneDecimal(humidity),
		PrecipitationMM: roundToOneDecimal(precipMM),
		WindSpeedMph:    roundToOneDecimal(windMph),
		Timestamp:       time.Now(),
	}, nil
}

func roundToOneDecimal(f float64) float64 {
	return float64(int(f*10.0+0.5)) / 10.0
}
