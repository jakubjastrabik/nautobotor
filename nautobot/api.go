package nautobot

import (
	"encoding/json"
	"log"
)

type Results struct {
	Family   Family `json:"family"`
	Address  string `json:"address"`
	Status   Status `json:"status"`
	Dns_name string `json:"dns_name"`
}

// IPaddress is structure for pars webhook intput data
type APIIPaddress struct {
	Count   int `json:"count"`
	Event   string
	Results []Results `json:"results"`
}

// NewIPaddress Unmarshal input byte to json struct
func NewAPIaddress(payload []byte) *APIIPaddress {
	var ip_add APIIPaddress
	ip_add.Event = "created"

	err := json.Unmarshal(payload, &ip_add)
	if err != nil {
		log.Println(err)
	}

	return &ip_add
}
