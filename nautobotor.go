package nautobotor

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/jakubjastrabik/nautobotor/ramrecords"
	"github.com/miekg/dns"
)

// Nautobotor is an nautobotor structure
type Nautobotor struct {
	WebAddress string
	RM         *ramrecords.RamRecord
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

// Name implements the Handler interface.
func (n Nautobotor) Name() string { return "nautobotor" }
