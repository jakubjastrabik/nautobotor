package nautobotor

import (
	"errors"
	"strings"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/miekg/dns"
)

// init registers this plugin.
func init() { plugin.Register("nautobotor", setup) }

// setup is the function that gets called when the config parser see the token "nautobotor". Setup is responsible
// for parsing any extra options the nautobotor plugin may have. The first token this function sees is "nautobotor".
func setup(c *caddy.Controller) error {

	nautobotorPlugin, err := newNautobotor(c)
	if err != nil {
		return plugin.Error("Nautobotor", err)
	}

	// Add a startup function that will -- after all plugins have been loaded -- check if the
	// prometheus plugin has been used - if so we will export metrics. We can only register
	// this metric once, hence the "once.Do".
	c.OnStartup(func() error {
		once.Do(func() {
			m := dnsserver.GetConfig(c).Handler("prometheus")
			if m == nil {
				return
			}
			if x, ok := m.(*metrics.Metrics); ok {
				x.MustRegister(requestCount)
			}
		})
		return nil
	})

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return nautobotorPlugin
	})

	// All OK, return a nil error.
	return nil
}

func newNautobotor(c *caddy.Controller) (*Nautobotor, error) {
	webaddress := ""

	log.Debug("Starting Nautobotor")

	re := New()

	re.zones = make([]string, 5)
	re.zones = []string{"lastmile.sk.", "if.lastmile.sk."}

	for _, zone := range re.zones {
		s := "test."
		ip := "192.168.1.212"
		ttl := "60"
		rr, err := dns.NewRR(s + zone + ttl + "A" + ip)
		if err != nil {
			return re, errors.New("Could not parse Nautobotor config")
		}

		rr.Header().Name = strings.ToLower(rr.Header().Name)
		re.m[zone] = append(re.m[zone], rr)
	}

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "webaddress":
					if !c.NextArg() {
						log.Error(c.ArgErr())
					}
					webaddress = c.Val()
				}
				if !c.Next() {
					break
				}
			}
		}
	}

	re.WebAddress = webaddress

	return re, nil
}
