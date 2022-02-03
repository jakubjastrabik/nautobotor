package ramrecords

import (
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

// AddZone handling proces to generate all necessary zone records wtih multiple types
func (re *RamRecord) AddZone(dnsName string, dnsNS map[string]string) {
	log.Debug("adding zone to zones array")
	zone := parseZone(dnsName)

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

// AddPTRZone handling proces to generate all necessary PTR zone records wtih multiple types
func (re *RamRecord) AddPTRZone(ipFamily int8, ip, dnsName string, dnsNS map[string]string) {
	log.Debug("adding PTR zone to zones array")

	zone := parsePTRzone(ipFamily, ip)

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

		re.handlePTRAddZone(zone, parseZone(dnsName), dnsNS)
	}
}

// RemoveRecord remove a record from zone
func (re *RamRecord) RemoveRecord(ipFamily int8, ip, dnsName string) {
	zone := parseZone(dnsName)

	switch ipFamily {
	case 4:
		// Delete A
		re.handleRemoveRecord(zone, "", strings.Split(dnsName, ".")[0]+" A "+cutCIDRMask(ip))
		// Delete PTR
		re.handleRemoveRecord(parseZone(dnsName), parsePTRzone(ipFamily, ip), createRe(ip)+" PTR "+strings.Split(dnsName, ".")[0])
	case 6:
		re.handleRemoveRecord(zone, "", strings.Split(dnsName, ".")[0]+" AAAA "+cutCIDRMask(ip))
		re.handleRemoveRecord(parseZone(dnsName), parsePTRzone(ipFamily, ip), createRe(ip)+" PTR "+strings.Split(dnsName, ".")[0])
	}

}

// AddRecord adds a record to the zone
func (re *RamRecord) AddRecord(ipFamily int8, ip, dnsName string) {
	log.Debug("adding record to the zone records array")

	// TODO: need to implement way to handle different types of DNS record
	switch ipFamily {
	case 4:
		// Add A
		re.newRecord(parseZone(dnsName), strings.Split(dnsName, ".")[0]+" A "+cutCIDRMask(ip))
	case 6:
		// Add AAAA
		re.newRecord(parseZone(dnsName), strings.Split(dnsName, ".")[0]+" AAAA "+cutCIDRMask(ip))
	}

	re.newPTRRecord(parseZone(dnsName), parsePTRzone(ipFamily, ip), createRe(ip)+" PTR "+strings.Split(dnsName, ".")[0])
}

// UpdateRecord update a record in the zone
func (re *RamRecord) UpdateRecord(ipFamily int8, ip, dnsName string) {
	log.Debug("updating record from the zone records array")

	// Prepare variables
	zone := parseZone(dnsName)
	ipO := cutCIDRMask(ip)

	for _, record := range re.M[zone] {
		log.Debug("try find the record in the zone")
		// try find the record in the zone
		if ipO == dns.Field(record, 1) {
			log.Debug("verify if record is correct")
			// verify if record is correct
			var s string
			dnsNameO := record.Header().Name
			zone := parseZone(dnsNameO)

			switch ipFamily {
			case 4:
				s = (strings.Split(dnsNameO, ".")[0] + " A " + cutCIDRMask(ip))
			case 6:
				s = (strings.Split(dnsNameO, ".")[0] + " AAAA " + cutCIDRMask(ip))
			}

			rr := handleCreateNewRR(zone, s)

			if dns.IsDuplicate(record, rr) {
				// If the record is corret => (deleted, create new one) = Update record
				log.Debug("delete record, creating new record")

				// remove existing record
				re.RemoveRecord(ipFamily, ip, dnsNameO)
				// add new record
				re.AddRecord(ipFamily, ip, dnsName)
			}

			return
		}
	}
}

// TODO: need to handle duplicated FQDN records
// May useful this function
// dns.IsDuplicate()

func InitRamRecords() (*RamRecord, error) {
	re := New()

	return re, nil
}
