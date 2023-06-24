package contracts

import "time"

type CreateCARequest struct {
	Name              string    `json:"name"`
	Organization      string    `json:"organization"`
	Country           string    `json:"country"`
	State             string    `json:"state"`
	Locality          string    `json:"locality"`
	PostalCode        string    `json:"postalCode"`
	StreetAddress     string    `json:"streetAddress"`
	Expiration        time.Time `json:"expiration"`
	ParentCA          string    `json:"parentCA"`
	ParentKeyPassword string    `json:"parentKeyPassword"`
	KeyID             string    `json:"key"`
	KeyPassword       string    `json:"keyPassword"`
}
