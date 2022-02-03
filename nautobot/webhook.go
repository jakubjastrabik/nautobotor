package nautobot

import (
	"encoding/json"
	"log"
)

type Family struct {
	Value int8 `json:"value"`
}

type Status struct {
	Value string `json:"value"`
}
type Data struct {
	Family   Family `json:"family"`
	Address  string `json:"address"`
	Status   Status `json:"status"`
	Dns_name string `json:"dns_name"`
}

// IPaddress is structure for pars webhook intput data
type IPaddress struct {
	Event string `json:"event"`
	Data  Data   `json:"data"`
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
