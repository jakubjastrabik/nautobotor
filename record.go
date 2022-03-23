package nautobotor

import (
	"net"
	"strings"

	"github.com/miekg/dns"
)

// handleCreateNewRR create new dnsRR
func handleCreateNewRR(zone, s string) dns.RR {
	rr, err := dns.NewRR("$ORIGIN " + zone + "\n" + s + "\n")
	if err != nil {
		log.Errorf("handleCreateNewRR() Error creating new record: err=%s\n", err)
		return nil
	}
	return rr
}

// cutCIDRMask Cut of CIDRMask from IP address
func cutCIDRMask(ip string) string {
	log.Debug("cutting of CIDRMask from IP address")

	ipvAddr, _, err := net.ParseCIDR(ip)
	if err != nil {
		log.Errorf("error parse IP address: err=%s\n", err)
	}
	return ipvAddr.String()
}

// createRRString Create string for dns.RR
// TODO(jakub): Need to create a way, to generate different DNS types like as CNAME
func createRRString(t, fqdn, ip string) string {

	switch t {
	case "IPv4":
		return strings.Split(fqdn, ".")[0] + " A " + cutCIDRMask(ip)
	case "IPv6":
		return strings.Split(fqdn, ".")[0] + " AAAA " + cutCIDRMask(ip)
	default:
		log.Errorf("createRRString() Undefined option in rr record string: %s", t)
		return ""
	}

}
