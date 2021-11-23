package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/juandes/cool-new-project/internal/client"
)

const (
	// The amount of days to look back for the average clicks
	daysToConsider = 30
)

// Handler function that computes the average clicks by country for Bitlinks in a user's default group.
// The function does three calls to Bitly API: first, it gets the user's info using the provided token.
// Second, it gets the Bitlinks belonging to the user's default group.
// Lastly, it gets the clicks by country for each Bitlink.
// After getting the data, it counts all the clicks by country and computes the average clicks by country in the last 30 days.
func computeAveragesByCountry(w http.ResponseWriter, r *http.Request) {
	// Get the token from the request
	token := r.URL.Query().Get("token")
	if token == "" {
		w.Write([]byte(fmt.Sprint("no token provided")))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// The client we will use to make the calls to the Bitly API.
	client := client.NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// In this map we will store the total clicks for each country
	totalClicksByCountry := make(map[string]int)

	// Get the user's info
	user, err := client.GetUser(ctx, token)
	if err != nil {
		w.Write([]byte(fmt.Sprint(err)))
		return
	}

	// Get the user's default group info.
	bitlinks, err := client.GetBitlinksByGroup(ctx, user.DefaultGroupGUID, token)
	if err != nil {
		w.Write([]byte(fmt.Sprint(err)))
	}

	// We will use this channel to send the clicks by country.
	vals := make(chan map[string]int, len(bitlinks))

	// Iterate over each bitlink to obtain their clicks by country.
	for _, bitlink := range bitlinks {
		go func(bitlink string) {
			cbc, err := client.GetClicksByCountry(ctx, bitlink, daysToConsider, token, vals)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			vals <- cbc
		}(bitlink)
	}

	// Track how many Bitlinks we have processed.
	bitlinksProcessed := 0

	for {
		select {
		case val := <-vals:
			for country, clicks := range val {
				totalClicksByCountry[country] += clicks
			}

			bitlinksProcessed++
			// Break when we have processed all the Bitlinks.
			if bitlinksProcessed >= len(bitlinks) {
				averages := calculateAverageClicks(totalClicksByCountry)

				// Convert the response to JSON and send it to the client.
				jsonResponse, err := json.Marshal(&AverageByCountryResponse{
					Averages: averages,
				})
				if err != nil {
					w.Write([]byte(fmt.Sprintf("error marshalling JSON response: %v", err)))
				}
				w.Write(jsonResponse)
				return
			}
		case <-time.After(1 * time.Second):
			w.Write([]byte(fmt.Sprintf("timeout")))
			return
		}
	}
}

func calculateAverageClicks(counts map[string]int) map[string]float64 {
	averages := make(map[string]float64)
	for country, count := range counts {
		averages[country] = float64(count) / float64(daysToConsider)
	}
	return averages
}
