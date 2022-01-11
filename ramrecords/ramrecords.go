package ramrecords

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

type RamRecord struct {
	Zones []string            // Array of zones
	M     map[string][]dns.RR // Map of DNS Records
}

// New returns a pointer to a new and intialized Records.
func New() *RamRecord {
	n := new(RamRecord)
	n.M = make(map[string][]dns.RR)
	return n
}

func NewRamRecords() (*RamRecord, error) {
	re := New()

	re.Zones = make([]string, 5)

	re.Zones = []string{"lastmile.sk.", "if.lastmile.sk."}

	for _, zone := range re.Zones {
		s := "test."
		ip := "192.168.1.1"
		ttl := 60
		rr, err := dns.NewRR(fmt.Sprintf("%s %d A %s", s+zone, ttl, ip))
		if err != nil {
			return re, errors.New("Could not parse Nautobotor config")
		}

		rr.Header().Name = strings.ToLower(rr.Header().Name)
		re.M[zone] = append(re.M[zone], rr)
	}
	return re, nil
}
