package nautobotor

import (
	"net"
	"strings"
	"time"

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

// createRRString Create string for dns.RR
// TODO(jakub): Need to create a way, to generate different DNS types like as CNAME
func createRRString(t, fqdn, ip string) string {

	switch t {
	case "IPv4":
		return strings.Split(fqdn, ".")[0] + " A " + cutCIDRMask(ip)
	case "IPv6":
		return strings.Split(fqdn, ".")[0] + " AAAA " + cutCIDRMask(ip)
	case "NS":
		return "@ NS " + strings.Split(fqdn, ".")[0]
	case "PTR":
		return createRe(ip) + " PTR " + fqdn
	case "PTRNS":
		return "@ NS " + fqdn
	case "SOA":
		return "@ SOA " + fqdn + ip + " dns-admin" + ip + " " + time.Now().Format("2006010215") + " 7200 3600 1209600 3600"
	default:
		log.Errorf("createRRString() Undefined option in rr record string: %s", t)
		return ""
	}

}

// parsePTRzone Create reverse ZONE
func parsePTRzone(ipFamily, ip string) string {
	// Convert IP to zone name
	newIP := cutCIDRMask(ip)

	if ipFamily == "IPv4" {
		newIP = newIP + "/24"
	} else {
		newIP = newIP + "/32"
	}

	_, zone, err := net.ParseCIDR(newIP)
	if err != nil {
		log.Error("failed to parse IP address")
	}

	dd, err := dns.ReverseAddr(zone.IP.String())
	dd = strings.Replace(dd, "0.", "", 1)
	if err != nil {
		log.Debugf("Issue generate ReverseAddr error= %s", err)
	}

	return dd
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

// createRe Create reverse ADDPREESS
func createRe(ip string) string {
	a, err := dns.ReverseAddr(cutCIDRMask(ip))
	if err != nil {
		log.Debugf("Issue generate ReverseAddr error= %s", err)
	}
	return a
}
