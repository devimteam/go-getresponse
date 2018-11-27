package getresponse

// Campaign holds the representation of a campaign
type Campaign struct {
	CampaignID string  `json:"campaignId"` // required
	Name       string  `json:"name,omitempty"`
	Href       *string `json:"href,omitempty"`
}

// CustomField holds key value sets
type CustomField struct {
	CustomFieldID string   `json:"customFieldId"`
	Value         []string `json:"value"`
	Href          *string  `json:"href,omitempty"`
}

// Geolocation holds geo data on contacts
type Geolocation struct {
	Latitude      *string `json:"latitude,omitempty"`
	Longitude     *string `json:"longitude,omitempty"`
	ContinentCode *string `json:"continentCode,omitempty"`
	CountryCode   *string `json:"countryCode,omitempty"`
	Region        *string `json:"region,omitempty"`
	PostalCode    *string `json:"postalCode,omitempty"`
	DmaCode       *string `json:"dmaCode,omitempty"`
	City          *string `json:"city,omitempty"`
}

type Tag struct {
	TagID string `json:"tagId"`
}

// Contact represents a GR contact
type Contact struct {
	ContactID         *string       `json:"contactId,omitempty"`
	Href              *string       `json:"href,omitempty"`
	Name              *string       `json:"name,omitempty"`
	Email             *string       `json:"email,omitempty"`
	Note              *string       `json:"note,omitempty"`
	DayOfCycle        *int32        `json:"dayOfCycle,omitempty"`
	Origin            *string       `json:"origin,omitempty"`
	CreatedOn         *string       `json:"createdOn,omitempty"` // there doesn't seem to be any docs on what timezone these times are in so I'm leaving them as strings (timeZone below is the user's timezone)
	ChangedOn         *string       `json:"changedOn,omitempty"`
	Campaign          *Campaign     `json:"campaign,omitempty"`
	Geolocation       *Geolocation  `json:"geolocation,omitempty"`
	Tags              []Tag         `json:"tags,omitempty"`
	CustomFieldValues []CustomField `json:"customFieldValues,omitempty"`
	TimeZone          *string       `json:"timeZone,omitempty"`
	IPAddress         *string       `json:"ipAddress,omitempty"`
	Activities        *string       `json:"activities,omitempty"`
	Scoring           *int64        `json:"scoring,omitempty"`
}
