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

func (re *RamRecord) addRecord(ipFamily int8, ip, zone, dnsName string) {
	log.Debug("adding record to the zone records array")

	// TODO: need to implement way to handle different types of DNS record
	switch ipFamily {
	case 4:
		re.newRecord(zone, dnsName+" A "+cutCIDRMask(ip))
	case 6:
		re.newRecord(zone, dnsName+" AAAA "+cutCIDRMask(ip))
	}
}

// TODO: need to handle duplicated FQDN records
// May useful this function
// dns.IsDuplicate()

func InitRamRecords() (*RamRecord, error) {
	re := New()

	// Test static string
	// TODO: replace with dynamic variables gether from nautobot
	dnsNS := map[string]string{
		"ans-m1": "172.16.5.90",
		"arn-t1": "172.16.5.76",
		"arn-x1": "172.16.5.77",
	}
	// Test static string
	// TODO: replace with dynamic variables gether from nautobot
	ipFamily := int8(4)
	ip_add := "192.168.1.1/24"
	dnsName := "test.if.lastmile.sk"

	// Add zone to Zones
	re.addZone(parseZone(dnsName), dnsNS)

	// Add record to zone table
	re.addRecord(ipFamily, ip_add, parseZone(dnsName), strings.Split(dnsName, ".")[0])

	return re, nil
}
