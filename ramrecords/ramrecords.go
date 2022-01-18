package ramrecords

import (
	"net"
	"strings"
	"time"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
)

type RamRecord struct {
	Zones []string            // Array of zones
	M     map[string][]dns.RR // Map of DNS Records
}

var log = clog.NewWithPlugin("nautobotor")

// NewRamRecords is used to initialize space for all records
// allocated first sets of records from nautobot via api.
// Returns a pointer to a new and intialized Records.
func NewRamRecords() (*RamRecord, error) {
	log.Debug("initializing ramrecords array")
	re := new(RamRecord)
	re.M = make(map[string][]dns.RR)

	return re, nil
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

func (re *RamRecord) AddZone(zone string) (*RamRecord, error) {
	log.Debug("Start adding zone to ramrecords")
	re.Zones = append(re.Zones, zone)

	// TODO: auto generate this section from the nautobot api response
	// soa, create a new SOA record
	re.newRecord(zone, "@ SOA ns.if.lastmile.sk. noc-srv.lastmile.sk. "+time.Now().Format("2006010215")+" 7200 3600 1209600 3600")

	dnsServer := map[string]string{
		"ans-m1": "172.16.5.90",
		"arn-t1": "172.16.5.76",
		"arn-x1": "172.16.5.77",
	}

	// TODO: auto generate this section from the nautobot api response
	// NS, create a new NS record
	for k, v := range dnsServer {
		re.newRecord(zone, "@ NS "+k)
		re.newRecord(zone, k+" A "+v)
	}

	return re, nil
}

func (re *RamRecord) AddRecord(ipFamily int8, ip string, dnsName string, zone string) (*RamRecord, error) {
	log.Debug("Start adding record to ramrecords")

	// Cut of CIDRMask
	ipvAddr, _, err := net.ParseCIDR(ip)
	if err != nil {
		log.Errorf("error parse IP address: err=%s\n", err)
	}

	switch ipFamily {
	case 4:
		re.newRecord(zone, dnsName+" A "+ipvAddr.String())
	case 6:
		re.newRecord(zone, dnsName+" AAAA "+ipvAddr.String())
	}

	log.Debug("After Record procesing procesing", re.M[zone])

	return re, nil
}
