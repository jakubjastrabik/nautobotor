package nautobotor

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/reuseport"
	"github.com/coredns/coredns/plugin/transfer"
	"github.com/coredns/coredns/request"
	"github.com/jakubjastrabik/nautobotor/nautobot"
	"github.com/miekg/dns"
)

// Nautobotor is an nautobotor structure
type Nautobotor struct {
	WebAddress  string
	NautobotURL string
	Token       string
	NS          map[string]string

	Zones
	transfer *transfer.Transfer

	ln   net.Listener
	mux  *http.ServeMux
	Next plugin.Handler
}

// Zones maps zone names to a *Zone.
type Zones struct {
	Z     map[string]*Zone // A map mapping zone (origin) to the Zone's data
	Names []string         // All the keys from the map Z as a string slice
}

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("nautobotor")

// ServeDNS implements the plugin.Handler interface. This method gets called when nautobotor is used
// in a Server.
func (n Nautobotor) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()
	zone := plugin.Zones(n.Zones.Names).Matches(qname)

	if zone == "" {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}

	z, ok := n.Zones.Z[zone]
	if !ok || z == nil {
		return dns.RcodeServerFailure, nil
	}

	// If transfer is not loaded, we'll see these, answer with refused (no transfer allowed).
	if state.QType() == dns.TypeAXFR || state.QType() == dns.TypeIXFR {
		return dns.RcodeRefused, nil
	}

	// This is only for when we are a secondary zones.
	if r.Opcode == dns.OpcodeNotify {
		if z.isNotify(state) {
			m := new(dns.Msg)
			m.SetReply(r)
			m.Authoritative = true
			w.WriteMsg(m)

			log.Infof("Notify from %s for %s: checking transfer", state.IP(), zone)
			ok, err := z.shouldTransfer()
			if ok {
				z.TransferIn()
			} else {
				log.Infof("Notify from %s for %s: no SOA serial increase seen", state.IP(), zone)
			}
			if err != nil {
				log.Warningf("Notify from %s for %s: failed primary check: %s", state.IP(), zone, err)
			}
			return dns.RcodeSuccess, nil
		}
		log.Infof("Dropping notify from %s for %s", state.IP(), zone)
		return dns.RcodeSuccess, nil
	}

	z.RLock()
	exp := z.Expired
	z.RUnlock()
	if exp {
		log.Errorf("Zone %s is expired", zone)
		return dns.RcodeServerFailure, nil
	}

	answer, ns, extra, result := z.Lookup(ctx, state, qname)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer, m.Ns, m.Extra = answer, ns, extra

	switch result {
	case Success:
	case NoData:
	case NameError:
		m.Rcode = dns.RcodeNameError
	case Delegation:
		m.Authoritative = false
	case ServerFailure:
		// If the result is SERVFAIL and the answer is non-empty, then the SERVFAIL came from an
		// external CNAME lookup and the answer contains the CNAME with no target record. We should
		// write the CNAME record to the client instead of sending an empty SERVFAIL response.
		if len(m.Answer) == 0 {
			return dns.RcodeServerFailure, nil
		}
		//  The rcode in the response should be the rcode received from the target lookup. RFC 6604 section 3
		m.Rcode = dns.RcodeServerFailure
	}

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil

}

// getApiData send api GET request to nautobot
// return data
func (n *Nautobotor) getApiData() error {

	req, err := http.NewRequest("GET", n.NautobotURL, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	req.Header.Set("Authorization", "Token "+n.Token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error on response err=%s\n", err)
		return err
	}
	defer resp.Body.Close()

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error while reading the response bytes err=%s\n", err)
		return err
	}

	// Unmarshal data to strcut
	log.Debug(nautobot.NewAPIaddress(payload))
	err = n.handleAPIData(nautobot.NewAPIaddress(payload))
	if err != nil {
		log.Errorf("error handling DNS data: err=%s\n", err)
		return err
	}

	return nil

}

// onStartup handling web request and response
// used for reciving webhook
// used for sending GET to nautobot API, to get all IP addresses
func (n *Nautobotor) onStartup() error {
	ln, err := reuseport.Listen("tcp", n.WebAddress)
	if err != nil {
		return err
	}

	n.ln = ln
	n.mux = http.NewServeMux()

	n.mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Start handling webhook data")

		// handleWebhook are used to processed nautobot webhook
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("error reading request body: err=%s\n", err)
			return
		}
		defer r.Body.Close()

		// Unmarshal data to strcut
		err = n.handleData(nautobot.NewIPaddress(payload))
		if err != nil {
			log.Errorf("error handling DNS data: err=%s\n", err)
		}

	})

	go func() {
		err := http.Serve(n.ln, n.mux)
		if err != nil {
			log.Errorf("errro initializing web server: err=%s\n", err)
		}
	}()

	return nil
}

// handleData are used to handle incoming data structures
// returning pointers to nautobot DNS records structures
func (n *Nautobotor) handleAPIData(ip *nautobot.APIIPaddress) error {
	log.Debug("Start handling DNS record")
	log.Debug("Unmarshaled data from webhook to be added to DNS: data=", ip)

	switch ip.Event {

	case "created":
		log.Debug("Received API data to creat")
		for _, i := range ip.Results {
			dnsName := parseZone(i.Dns_name)

			// Handle zone
			if err := n.Zones.AddZone(dnsName, ""); err != nil {
				log.Errorf("handleApiData() Error creating zone: %s, err=%s\n", dnsName, err)
			}

			if n.Zones.Z[dnsName].Apex.NS == nil {
				// Handle Add zone NS record
				for v := range n.NS {
					if err := n.Zones.Z[dnsName].Insert(handleCreateNewRR(dnsName, createRRString("NS", v, ""))); err != nil {
						log.Errorf("handleApiData() Unable add NS record to the zone: %s error = %s\n", err, dnsName)
					}
				}
			}

			// Handel Add zone records
			if err := n.Zones.Z[dnsName].Insert(handleCreateNewRR(dnsName, createRRString(i.Family.Label, i.Dns_name, i.Address))); err != nil {
				log.Errorf("handleApiData() Unable add record to the zone: %s error = %s\n", err, dnsName)
			}

			// Handle Add PTR zone
			ptrZone := parsePTRzone(i.Family.Label, i.Address)

			// Handle PTR zone
			if err := n.Zones.AddZone(ptrZone, dnsName); err != nil {
				log.Errorf("handleApiData() Error creating zone: %s, err=%s\n", ptrZone, err)
			}

			if n.Zones.Z[ptrZone].Apex.NS == nil {
				// Handle Add zone NS record
				for k, v := range n.NS {
					if err := n.Zones.Z[ptrZone].Insert(handleCreateNewRR(ptrZone, createRRString("PTRNS", k+"."+dnsName, ""))); err != nil {
						log.Errorf("handleApiData() Unable add NS record to the zone: %s error = %s\n", err, dnsName)
					}
					// ip in ptr fqdn
					if err := n.Zones.Z[ptrZone].Insert(handleCreateNewRR(i.Dns_name, createRRString("PTR", k+"."+dnsName, v))); err != nil {
						log.Errorf("handleApiData() Unable add NS record to the zone: %s error = %s\n", err, dnsName)
					}
				}
			}

			// Add record to the zone
			if err := n.Zones.Z[ptrZone].Insert(handleCreateNewRR(i.Dns_name, createRRString("PTR", i.Dns_name+".", i.Address))); err != nil {
				log.Errorf("handleApiData() Unable add NS record to the zone: %s error = %s\n", err, dnsName)
			}

		}
	default:
		log.Errorf("Unable processed Event: %v", ip.Event)
	}

	return nil
}

// handleData are used to handle incoming data structures
// returning pointers to nautobot DNS records structures
func (n *Nautobotor) handleData(ip *nautobot.IPaddress) error {
	log.Debug("Start handling DNS record")
	log.Debug("Unmarshaled data from webhook to be add to DNS: data=", ip)

	dnsName := parseZone(ip.Data.Dns_name)

	switch ip.Event {
	case "created":
		log.Debug("Received webhook to creat")

		// Handle Normal zone
		if err := n.Zones.AddZone(dnsName, ""); err != nil {
			log.Errorf("handleData() Error creating zone: %s, err=%s\n", dnsName, err)
		}

		// Handle Add zone NS record
		for v := range n.NS {
			if err := n.Zones.Z[dnsName].Insert(handleCreateNewRR(dnsName, createRRString("NS", v, ""))); err != nil {
				log.Errorf("handleApiData() Unable add NS record to the zone: %s error = %s\n", err, dnsName)
			}
		}

		// Add record to the zone
		if err := n.Zones.Z[dnsName].Insert(handleCreateNewRR(dnsName, createRRString(ip.Data.Family.Label, ip.Data.Dns_name, ip.Data.Address))); err != nil {
			log.Errorf("handleData() Unable add record to the zone: %s error = %s\n", err, dnsName)
		}

		// 	// Handle PTR zones
		// 	n.RM.AddPTRZone(ip.Data.Family.Value, ip.Data.Address, ip.Data.Dns_name, n.NS)

	// 	n.RM.AddRecord(ip.Data.Family.Value, ip.Data.Address, ip.Data.Dns_name)

	// case "deleted":
	// 	log.Debug("Received webhook to delet")
	// 	// Remove record from the zone
	// 	n.RM.RemoveRecord(ip.Data.Family.Value, ip.Data.Address, ip.Data.Dns_name)

	// case "updated":
	// 	log.Debug("Received webhook to update")
	// 	// Update record in the zone
	// 	n.RM.UpdateRecord(ip.Data.Family.Value, ip.Data.Address, ip.Data.Dns_name, n.NS)
	default:
		log.Errorf("Unable processed Event: %v", ip.Event)
	}

	return nil
}

// Name implements the Handler interface.
func (n Nautobotor) Name() string { return "nautobotor" }
