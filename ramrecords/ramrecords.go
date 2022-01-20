package ramrecords

import (
	"errors"
	"fmt"
	"strings"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
)

type RamRecord struct {
	Zones []string            // Array of zones
	M     map[string][]dns.RR // Map of DNS Records
}

// Init log variable
var log = clog.NewWithPlugin("nautobotor")

// New returns a pointer to a new and intialized Records.
func New() *RamRecord {
	log.Debug("initializing RamRecord struct")
	n := new(RamRecord)
	n.M = make(map[string][]dns.RR)
	return n
}

func (r *RamRecord) addZone(zone string) {
	log.Debug("adding zone to zones array")

	// If zone is empty
	if r.Zones == nil {
		r.Zones = make([]string, 1)
		r.Zones = []string{zone}
	} else {
		// If zone already exists
		for _, z := range r.Zones {
			if z == zone {
				return
			}
		}
		// If not, add zone to the struct
		r.Zones = append(r.Zones, zone)
	}
}

func InitRamRecords() (*RamRecord, error) {
	re := New()
	// re.Zones = make([]string, 1)

	re.addZone("if.lastmile.sk.")
	re.addZone("if.lastmile.sk.")

	// re.Zones = []string{"if.lastmile.sk."}

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
