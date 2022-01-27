package nautobotor

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/reuseport"
	"github.com/coredns/coredns/request"
	"github.com/jakubjastrabik/nautobotor/nautobot"
	"github.com/jakubjastrabik/nautobotor/ramrecords"
	"github.com/miekg/dns"
)

// Nautobotor is an nautobotor structure
type Nautobotor struct {
	WebAddress string
	RM         *ramrecords.RamRecord
	ln         net.Listener
	mux        *http.ServeMux
	Next       plugin.Handler
}

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("nautobotor")

// ServeDNS implements the plugin.Handler interface. This method gets called when nautobotor is used
// in a Server.
func (n Nautobotor) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()
	zone := plugin.Zones(n.RM.Zones).Matches(qname)

	if zone == "" {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}

	// New we should have some data for this zone, as we just have a list of RR, iterate through them, find the qname
	// and see if the qtype exists. If so reply, if not do the normal DNS thing and return either A or AAAA.
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	nxdomain := true
	var soa dns.RR
	for _, r := range n.RM.M[zone] {
		if r.Header().Rrtype == dns.TypeSOA && soa == nil {
			soa = r
		}
		if r.Header().Name == qname {
			nxdomain = false
			if r.Header().Rrtype == state.QType() {
				m.Answer = append(m.Answer, r)
			}
		}
	}

	// handle nxdomain, NODATA and normal response here.
	if nxdomain {
		m.Rcode = dns.RcodeNameError
		if soa != nil {
			m.Ns = []dns.RR{soa}
		}
		err := w.WriteMsg(m)
		if err != nil {
			log.Error(err)
		}
		return dns.RcodeSuccess, nil
	}

	if len(m.Answer) == 0 {
		if soa != nil {
			m.Ns = []dns.RR{soa}
		}
	}

	// Export metric with the server label set to the current server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	err := w.WriteMsg(m)
	if err != nil {
		log.Error(err)
	}
	return dns.RcodeSuccess, nil

}

// TODO: Add comment
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
func (n *Nautobotor) handleData(ip *nautobot.IPaddress) error {
	log.Debug("Start handling DNS record")
	log.Debug("Unmarshaled data from webhook to be add to DNS: data=", ip)

	switch ip.Event {
	case "created":
		log.Debug("Received webhook to creat")
		n.RM.AddRecord(ip.Data.Family.Value, ip.Data.Address, ip.Data.Dns_name)
	case "deleted":
		log.Debug("Received webhook to delet")
		n.RM.RemoveRecord(ip.Data.Family.Value, ip.Data.Address, ip.Data.Dns_name)
	case "edited":
		log.Debug("Received webhook to edit")
	default:
		log.Errorf("Unable processed Event: %v", ip.Event)
	}

	return nil
}

// Name implements the Handler interface.
func (n Nautobotor) Name() string { return "nautobotor" }
