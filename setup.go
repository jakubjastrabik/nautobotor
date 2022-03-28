package nautobotor

import (
	"errors"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/plugin/transfer"
)

var Version = "v0.6.0"

// init registers this plugin.
func init() { plugin.Register("nautobotor", setup) }

// setup is the function that gets called when the config parser see the token "nautobotor". Setup is responsible
// for parsing any extra options the nautobotor plugin may have. The first token this function sees is "nautobotor".
func setup(c *caddy.Controller) error {

	nautobotorPlugin, err := newNautobotor(c)
	if err != nil {
		return plugin.Error("Nautobotor", err)
	}

	// Log plugin version
	log.Infof("Started plugin version %s", Version)

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

	// Start handle zone transfer
	c.OnStartup(func() error {
		t := dnsserver.GetConfig(c).Handler("transfer")
		if t == nil {
			return nil
		}
		nautobotorPlugin.transfer = t.(*transfer.Transfer) // if found this must be OK.
		go func() {
			for _, n := range nautobotorPlugin.Zones.Names {
				nautobotorPlugin.transfer.Notify(n)
			}
		}()
		return nil
	})

	c.OnRestartFailed(func() error {
		t := dnsserver.GetConfig(c).Handler("transfer")
		if t == nil {
			return nil
		}
		go func() {
			for _, n := range nautobotorPlugin.Zones.Names {
				nautobotorPlugin.transfer.Notify(n)
			}
		}()
		return nil
	})

	// Download all zone data from nautobotor API
	c.OnStartup(func() error {
		err := nautobotorPlugin.getApiData()
		if err != nil {
			log.Errorf("Unable startup web server: err=%s\n", err)
			return err
		}
		return nil
	})

	// Listen for webhook update from nautobot
	c.OnStartup(func() error {
		err := nautobotorPlugin.onStartup()
		if err != nil {
			log.Errorf("Unable startup web server: err=%s\n", err)
			return err
		}
		return nil
	})

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		nautobotorPlugin.Next = next
		return nautobotorPlugin
	})

	// All OK, return a nil error.
	return nil
}

func newNautobotor(c *caddy.Controller) (Nautobotor, error) {
	var n = Nautobotor{}

	//	Parse the input data
	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "webaddress":
					if !c.NextArg() {
						log.Error(c.ArgErr())
					}
					n.WebAddress = c.Val()
				case "nautoboturl":
					if !c.NextArg() {
						log.Error(c.ArgErr())
					}
					n.NautobotURL = c.Val()

				case "token":
					if !c.NextArg() {
						log.Error(c.ArgErr())
					}
					n.Token = c.Val()
				}

				if !c.Next() {
					break
				}
			}
		}
	}
	if n.WebAddress == "" || n.NautobotURL == "" || n.Token == "" {
		return Nautobotor{}, errors.New("Could not parse config, or some input data are missing")
	}

	n.NS = map[string]string{
		"ans-m1": "172.16.5.90/24",
		"arn-t1": "172.16.5.76/24",
		"arn-x1": "172.16.5.77/24",
	}

	return n, nil
}
