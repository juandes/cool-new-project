package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// StatsService is a service that provides the average clicks by country for a user's default group.
type StatsService struct {
	server *http.Server
	// The token to use for the API calls. This is just for testing purposes!
	token string
	// The amount of days to look back for the average clicks
	daysToConsider int
}

// NewStatsService creates a new StatsService
func NewStatsService() *StatsService {
	return &StatsService{
		server: &http.Server{
			Addr: ":8080",
		},
		daysToConsider: 30,
	}
}

func (s *StatsService) Start() {
	http.HandleFunc("/averages", s.computeAveragesByCountry)
	err := s.server.ListenAndServe()
	// Let's just panic here.
	panic(err)
}

// handler function that computes the average clicks by country for Bitlinks in a user's default group.
// The function does three calls to Bitly API: first, it gets the user's info using the provided token.
// Second, it gets the Bitlinks belonging to the user's default group.
// Lastly, it gets the clicks by country for each Bitlink.
// After getting the data, it counts all the clicks by country and computes the average clicks by country in the last 30 days.
func (s *StatsService) computeAveragesByCountry(w http.ResponseWriter, r *http.Request) {
	// Setting the token here is dirty and hacky but
	// I'll do it for this small project.
	s.token = r.URL.Query().Get("token")

	client := &http.Client{}
	w.Header().Set("Content-Type", "application/json")

	// In this map we will store the total clicks for each country
	totalClicksByCountry := make(map[string]int)

	// Get the user's info
	user, err := s.getUser(client)
	if err != nil {
		w.Write([]byte(fmt.Sprint(err)))
	}

	// Get the user's default group info.
	bitlinks, err := s.getBitlinksByGroup(client, user.DefaultGroupGUID)
	if err != nil {
		w.Write([]byte(fmt.Sprint(err)))
	}

	// We will use this channel to send the clicks by country.
	vals := make(chan map[string]int, len(bitlinks))
	errc := make(chan error, 1)
	done := make(chan bool, 1)

	// Iterate over each bitlink to obtain their clicks by country.
	for _, bitlink := range bitlinks {
		go func(bitlink string) {
			cbc := s.getClicksByCountry(client, bitlink, vals, errc)
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
				close(done)
			}
		case err := <-errc:
			// Return if there's an error.
			w.Write([]byte(err.Error()))
			return
		case <-done:
			averages := s.calculateAverageClicks(totalClicksByCountry)

			// Convert the response to JSON and send it to the client.
			jsonResponse, err := json.Marshal(&AverageByCountryResponse{
				Averages: averages,
			})
			if err != nil {
				w.Write([]byte(fmt.Sprintf("error marshalling JSON response: %v", err)))
			}
			w.Write(jsonResponse)
			return
		case <-time.After(1 * time.Second):
			w.Write([]byte(fmt.Sprintf("timeout")))
			return
		}
	}
}
func (s *StatsService) doRequest(client *http.Client, endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error setting authorization: %v", err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	return bodyText, nil
}

func (s *StatsService) getUser(client *http.Client) (*User, error) {
	// Get the user's info.
	bodyText, err := s.doRequest(client, "https://api-ssl.bitly.com/v4/user")
	if err != nil {
		return nil, fmt.Errorf("error calling user's endpoint %v", err)
	}

	var user *User
	if err := json.Unmarshal(bodyText, &user); err != nil {
		return nil, fmt.Errorf("error unmarshalling User JSON: %v", err)
	}

	return user, nil
}
func (s *StatsService) getBitlinksByGroup(client *http.Client, defaultGroup string) ([]string, error) {
	// Here we will store all the bitlink's belonging to the given group.
	bitlinks := make([]string, 0)
	currentPage := 1

	// Iterate over all pages of Bitlinks.
	for {
		bodyText, err := s.doRequest(client, fmt.Sprintf("https://api-ssl.bitly.com/v4/groups/%s/bitlinks?page=%d", defaultGroup, currentPage))
		if err != nil {
			return nil, fmt.Errorf("error calling group's bitlinks endpoint %v", err)
		}

		var bitlinksByGroup *BitlinksByGroup
		if err := json.Unmarshal(bodyText, &bitlinksByGroup); err != nil {
			return nil, fmt.Errorf("error unmarshalling BitlinksByGroup JSON: %v", err)
		}

		for _, link := range bitlinksByGroup.Links {
			bitlinks = append(bitlinks, link.ID)
		}

		// We break if we have no more Bitlinks
		if currentPage == bitlinksByGroup.Pagination.Total {
			break
		}
		currentPage++
	}

	return bitlinks, nil
}

// Get the bitlink's clicks by country
func (s *StatsService) getClicksByCountry(client *http.Client, bitlink string, counts chan<- map[string]int, errc chan<- error) map[string]int {
	cbc := make(map[string]int)

	bodyText, err := s.doRequest(client, fmt.Sprintf("https://api-ssl.bitly.com/v4/bitlinks/%s/countries?unit=day&units=%d", bitlink, s.daysToConsider))
	if err != nil {
		errc <- fmt.Errorf("error calling group's bitlinks endpoint %v", err)
	}

	var metricsByCountry *MetricsByCountry
	if err := json.Unmarshal(bodyText, &metricsByCountry); err != nil {
		errc <- fmt.Errorf("error unmarshalling MetricsByCountry JSON: %v", err)
	}

	for _, metric := range metricsByCountry.Metrics {
		cbc[metric.Value] += metric.Clicks
	}

	return cbc
}

func (s *StatsService) calculateAverageClicks(counts map[string]int) map[string]float64 {
	averages := make(map[string]float64)
	for country, count := range counts {
		averages[country] = float64(count) / float64(s.daysToConsider)
	}
	return averages
}
