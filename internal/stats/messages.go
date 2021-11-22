package stats

// User is the response from Bitly's getUser endpoint
type User struct {
	Created      string `json:"created"`
	Modified     string `json:"modified"`
	Login        string `json:"login"`
	IsActive     bool   `json:"is_active"`
	Is2FaEnabled bool   `json:"is_2fa_enabled"`
	Name         string `json:"name"`
	Emails       []struct {
		Email      string `json:"email"`
		IsPrimary  bool   `json:"is_primary"`
		IsVerified bool   `json:"is_verified"`
	} `json:"emails"`
	IsSsoUser        bool   `json:"is_sso_user"`
	DefaultGroupGUID string `json:"default_group_guid"`
}

// BitlinksByGroup is the response from Bitly's "bitlinks by group" endpoint
type BitlinksByGroup struct {
	Links []struct {
		CreatedAt      string   `json:"created_at"`
		ID             string   `json:"id"`
		Link           string   `json:"link"`
		CustomBitlinks []string `json:"custom_bitlinks"`
		LongURL        string   `json:"long_url"`
		Title          string   `json:"title"`
		Archived       bool     `json:"archived"`
		CreatedBy      string   `json:"created_by"`
		ClientID       string   `json:"client_id"`
		Tags           []string `json:"tags"`
		Deeplinks      []string `json:"deeplinks"`
		References     struct {
			Group string `json:"group"`
		} `json:"references"`
	} `json:"links"`
	Pagination struct {
		Prev  string `json:"prev"`
		Next  string `json:"next"`
		Size  int    `json:"size"`
		Page  int    `json:"page"`
		Total int    `json:"total"`
	} `json:"pagination"`
}

// MetricsByCountry is the response from Bitly's "metrics by country" endpoint
type MetricsByCountry struct {
	UnitReference string `json:"unit_reference"`
	Metrics       []struct {
		Value  string `json:"value"`
		Clicks int    `json:"clicks"`
	} `json:"metrics"`
	Units int    `json:"units"`
	Unit  string `json:"unit"`
	Facet string `json:"facet"`
}

// AverageByCountryResponse is the response object.
type AverageByCountryResponse struct {
	// Where key is the country code and value is the average clicks.
	Averages map[string]float64 `json:"averages"`
}
