package ramrecords

import (
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// newRecord, generate dns.RR records for each zones, records
// data will be written to the ramRecord struct
func (re *RamRecord) newRecord(zone, s string) {
	log.Debug("handling dns record creation")

	rr, err := dns.NewRR("$ORIGIN " + zone + "\n" + s + "\n")
	if err != nil {
		log.Errorf("error creating new record: err=%s\n", err)
	}
	rr.Header().Name = strings.ToLower(rr.Header().Name)
	re.M[zone] = append(re.M[zone], rr)

	log.Debugf("Create newRecord: zone=%s, record=%s", zone, rr)
}

// handled zone, trying minimalized needs of code line
func (re *RamRecord) handleAddZone(zone string, dnsNS map[string]string) {
	log.Debug("handling zone creation")

	// Generate zone SOA record
	re.newRecord(zone, "@ SOA ns noc-srv.lastmile.sk. "+time.Now().Format("2006010215")+" 7200 3600 1209600 3600")

	// Generate NS record for zone
	for k, v := range dnsNS {
		re.newRecord(zone, "@ NS "+k)
		re.newRecord(zone, k+" A "+v)
	}
}

// Cut of CIDRMask from IP address
func cutCIDRMask(ip string) string {
	log.Debug("cutting of CIDRMask from IP address")

	ipvAddr, _, err := net.ParseCIDR(ip)
	if err != nil {
		log.Errorf("error parse IP address: err=%s\n", err)
	}
	return ipvAddr.String()
}

// Cut zone from FQDN
func parseZone(name string) string {
	name = strings.Trim(name, strings.Split(name, ".")[0])
	name = strings.Trim(name, ".") + "."
	return name
}

// createRe Create reverse ADDPREESS
func createRe(ip string) string {
	a, err := dns.ReverseAddr(cutCIDRMask(ip))
	if err != nil {
		log.Debugf("Issue generate ReverseAddr error= %s", err)
	}
	return a
}
