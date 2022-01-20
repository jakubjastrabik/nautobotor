package ramrecords

import (
	"errors"
	"fmt"
	"strings"
	"time"

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

func (re *RamRecord) newRecord(zone, s string) {
	rr, err := dns.NewRR("$ORIGIN " + zone + "\n" + s + "\n")
	if err != nil {
		log.Errorf("error creating new record: err=%s\n", err)
	}
	rr.Header().Name = strings.ToLower(rr.Header().Name)
	re.M[zone] = append(re.M[zone], rr)
	log.Debugf("Create newRecord: zone=%s, record=%s", zone, rr)
}

func (re *RamRecord) addZone(zone string) {
	log.Debug("adding zone to zones array")

	// If zone is empty
	if re.Zones == nil {
		re.Zones = make([]string, 1)
		re.Zones = []string{zone}
		// Generate zone SOA record
		re.newRecord(zone, "@ SOA ns.if.lastmile.sk. noc-srv.lastmile.sk. "+time.Now().Format("2006010215")+" 7200 3600 1209600 3600")
	} else {
		// If zone already exists
		for _, z := range re.Zones {
			if z == zone {
				return
			}
		}
		// If not, add zone to the struct
		re.Zones = append(re.Zones, zone)
	}
}

func InitRamRecords() (*RamRecord, error) {
	re := New()
	// re.Zones = make([]string, 1)

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
