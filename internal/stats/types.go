package stats

// AverageByCountryResponse is the response object.
type AverageByCountryResponse struct {
	// Where key is the country code and value is the average clicks.
	Averages map[string]float64 `json:"averages"`
}
