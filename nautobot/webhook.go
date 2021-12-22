package nautobot

import (
	"encoding/json"
	"log"
)

// IPaddress is structure for pars webhook intput data
type IPaddress struct {
	Event string `json:"event"`
	Data  struct {
		Family struct {
			Value int8 `json:"value"`
		} `json:"family"`
		Address string `json:"address"`
		Status  struct {
			Value string `json:"value"`
		} `json:"status"`
		Dns_name string `json:"dns_name"`
	} `json:"data"`
}

// NewIPaddress Unmarshal input byte to json struct
func NewIPaddress(payload []byte) *IPaddress {
	var ip_add IPaddress

	err := json.Unmarshal(payload, &ip_add)
	if err != nil {
		log.Println(err)
	}

	return &ip_add
}
