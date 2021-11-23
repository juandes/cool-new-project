package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	bitlyURL = "https://api-ssl.bitly.com/v4"
)

// Client provide access to bitlink API.
type Client struct {
	conn     *http.Client
	basePath string
}

func NewClient() *Client {
	return &Client{
		conn: &http.Client{
			Timeout: 5 * time.Second,
		},
		basePath: bitlyURL,
	}
}

// GetUser gets the user's information.
func (c *Client) GetUser(ctx context.Context, token string) (*User, error) {
	// Get the user's info.
	bodyText, err := c.doRequest(ctx, fmt.Sprintf("%s/user", c.basePath), token)
	if err != nil {
		return nil, fmt.Errorf("error calling user's endpoint %v", err)
	}

	var user *User
	if err := json.Unmarshal(bodyText, &user); err != nil {
		return nil, fmt.Errorf("error unmarshalling User JSON: %v", err)
	}

	return user, nil
}

// GetBitlinksByGroup gets the bitlinks from a group.
func (c *Client) GetBitlinksByGroup(ctx context.Context, defaultGroup, token string) ([]string, error) {
	// Here we will store all the bitlink's belonging to the given group.
	bitlinks := make([]string, 0)
	page := 1

	// Iterate over all pages of Bitlinks.
	for {
		bodyText, err := c.doRequest(ctx, fmt.Sprintf("%s/groups/%s/bitlinks?page=%d", c.basePath, defaultGroup, page), token)
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
		if page >= bitlinksByGroup.Pagination.Total {
			break
		}
		page++
	}

	return bitlinks, nil
}

// getClicksByCountry gets a Bitlink's clicks by country.
func (c *Client) GetClicksByCountry(ctx context.Context, bitlink string, units int, token string, counts chan<- map[string]int) (map[string]int, error) {
	clicksByCountry := make(map[string]int)

	bodyText, err := c.doRequest(ctx, fmt.Sprintf("%s/bitlinks/%s/countries?unit=day&units=%d", c.basePath, bitlink, units), token)
	if err != nil {
		return clicksByCountry, fmt.Errorf("error calling group's bitlinks endpoint %v", err)
	}

	var metricsByCountry *MetricsByCountry
	if err := json.Unmarshal(bodyText, &metricsByCountry); err != nil {
		return clicksByCountry, fmt.Errorf("error unmarshalling MetricsByCountry JSON: %v", err)
	}

	for _, metric := range metricsByCountry.Metrics {
		clicksByCountry[metric.Value] += metric.Clicks
	}

	return clicksByCountry, nil
}

func (c *Client) doRequest(ctx context.Context, endpoint, token string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := c.conn.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error setting authorization: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}
