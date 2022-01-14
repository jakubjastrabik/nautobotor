package nautobotor

import (
	"errors"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/jakubjastrabik/nautobotor/ramrecords"
)

var Version = "v0.2.4"

// init registers this plugin.
func init() { plugin.Register("nautobotor", setup) }

// setup is the function that gets called when the config parser see the token "nautobotor". Setup is responsible
// for parsing any extra options the nautobotor plugin may have. The first token this function sees is "nautobotor".
func setup(c *caddy.Controller) error {
	log.Debug("Start inicializing module configure values")

	nautobotorPlugin, err := newNautobotor(c)
	if err != nil {
		return plugin.Error("Nautobotor", err)
	}

	c.OnStartup(func() error {
		err := nautobotorPlugin.onStartup()
		if err != nil {
			log.Errorf("Unable startup web server: err=%s\n", err)
			return err
		}
		return nil
	})

	c.OnStartup(func() error {
		nautobotorPlugin.RM, err = ramrecords.NewRamRecords()
		if err != nil {
			log.Errorf("errro initializing module structure: err=%s\n", err)
			return err
		}
		return nil
	})
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

func newNautobotor(c *caddy.Controller) (Nautobotor, error) {
	var n = Nautobotor{}

	log.Debug("Starting parsing configuration values")

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "webaddress":
					if !c.NextArg() {
						log.Errorf("unable parsing configuration values: err=%s\n", c.ArgErr())
					}
					n.WebAddress = c.Val()
				}
				if !c.Next() {
					break
				}
			}
		}
	}

	if n.WebAddress == "" {
		return Nautobotor{}, errors.New("Could not parse webaddress from congiguration values")
	}

	log.Infof("Started plugin version %s", Version)
	return n, nil
}
