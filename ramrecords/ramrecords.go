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

func (re *RamRecord) AddZone(zone string) (*RamRecord, error) {
	log.Debug("Start adding zone to ramrecords")
	re.Zones = append(re.Zones, zone)

	// TODO: auto generate this section from the nautobot api response
	// soa, create a new SOA record
	soa, err := dns.NewRR(zone + " 60  IN SOA ns." + zone + " noc-srv.lastmile.sk. " + time.Now().Format("2006010215") + " 7200 3600 1209600 3600")
	if err != nil {
		log.Errorf("error creating SOA record: err=%s\n", err)
	}
	soa.Header().Name = strings.ToLower(soa.Header().Name)
	re.M[zone] = append(re.M[zone], soa)

	dnsServer := map[string]string{
		"ans-m1": "172.16.5.90",
		"arn-t1": "172.16.5.76",
		"arn-x1": "172.16.5.77",
	}

	// TODO: auto generate this section from the nautobot api response
	// NS, create a new NS record
	for k := range dnsServer {
		ns, err := dns.NewRR(zone + " 60  NS " + k + "." + zone)
		if err != nil {
			log.Errorf("error creating NS record: err=%s\n", err)
		}
		re.M[zone] = append(re.M[zone], ns)
	}

	// TODO: auto generate this section from the nautobot api response
	// a, create a new A record
	for k, v := range dnsServer {
		a, err := dns.NewRR(k + "." + zone + " 60  A " + v)
		if err != nil {
			log.Errorf("error creating A record: err=%s\n", err)
		}
		re.M[zone] = append(re.M[zone], a)
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
		a, err := dns.NewRR(dnsName + " 60  A " + ipvAddr.String())
		if err != nil {
			log.Errorf("error adding A record: err=%s\n", err)
		}
		re.M[zone] = append(re.M[zone], a)
	case 6:
		aaaa, err := dns.NewRR(dnsName + " 60  AAAA " + ipvAddr.String())
		if err != nil {
			log.Errorf("error adding AAAA record: err=%s\n", err)
		}
		re.M[zone] = append(re.M[zone], aaaa)
	}

	log.Debug("After Record procesing procesing", re.M[zone])

	return re, nil
}
