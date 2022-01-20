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

// newRecord, generate dns.RR records for each zones, records
// data will be written to the ramRecord struct
func (re *RamRecord) newRecord(zone, s string) {
	rr, err := dns.NewRR("$ORIGIN " + zone + "\n" + s + "\n")
	if err != nil {
		log.Errorf("error creating new record: err=%s\n", err)
	}
	rr.Header().Name = strings.ToLower(rr.Header().Name)
	re.M[zone] = append(re.M[zone], rr)
	log.Debugf("Create newRecord: zone=%s, record=%s", zone, rr)
}

// addZone handling proces to generate all necessary records wtih multiple types
func (re *RamRecord) addZone(zone string, dnsNS map[string]string) {
	log.Debug("adding zone to zones array")

	// If zone is empty
	if re.Zones == nil {
		re.Zones = make([]string, 1)
		re.Zones = []string{zone}

		re.handleAddZone(zone, dnsNS)
	} else {
		// If zone already exists
		for _, z := range re.Zones {
			if z == zone {
				return
			}
		}
		// If not, add zone to the struct
		re.Zones = append(re.Zones, zone)

		re.handleAddZone(zone, dnsNS)
	}
}

func InitRamRecords() (*RamRecord, error) {
	re := New()
	// re.Zones = make([]string, 1)

	dnsNS := map[string]string{
		"ans-m1": "172.16.5.90",
		"arn-t1": "172.16.5.76",
		"arn-x1": "172.16.5.77",
	}

	re.addZone("if.lastmile.sk.", dnsNS)

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
