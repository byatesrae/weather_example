package weather

// Summary represents weather datapoints for a location at a point in time.
type Summary struct {
	WindSpeed   float64 `json:"wind_speed"`          // The location windspeed in km/h.
	Temperature float64 `json:"temperature_degrees"` // The location temperature in degrees celsius.
}
