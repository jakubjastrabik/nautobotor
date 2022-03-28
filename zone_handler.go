package nautobotor

import (
	"strings"

	"github.com/coredns/coredns/plugin/pkg/upstream"
)

// AddZone checks if the zone exists if not, creating it
// returns nil, if already exists
func (z *Zones) AddZone(name, origin string) error {
	// Check if zone already exists
	if _, ok := z.Z[name]; !ok {
		// Check if Zones is initialized
		if z.Z == nil {
			z.Z = make(map[string]*Zone)
		}

		// Create new zone
		z.Z[name] = NewZone(name)
		z.Z[name].Upstream = upstream.New()

		if origin != "" {
			origin = "." + origin
			// insert soa to the zone
			if err := z.Z[name].Insert(handleCreateNewRR(name, createRRString("SOA", "ns", origin))); err != nil {
				log.Errorf("AddZone() Unable create SOA record, error = %s\n", err)
			}
		} else {
			// insert soa to the zone
			if err := z.Z[name].Insert(handleCreateNewRR(name, createRRString("SOA", "ns", ""))); err != nil {
				log.Errorf("AddZone() Unable create SOA record, error = %s\n", err)
			}
		}

		// Adding zone metadata to the list of Zones
		z.Names = append(z.Names, name)
	}
	return nil
}

// Cut zone from FQDN
func parseZone(name string) string {
	name = strings.Replace(name, strings.Split(name, ".")[0], "", 1)
	name = strings.Trim(name, ".") + "."
	return name
}
