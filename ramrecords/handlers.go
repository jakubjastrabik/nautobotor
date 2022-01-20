package ramrecords

import "time"

// handled zone, trying minimalized needs of code line
func (re *RamRecord) handleAddZone(zone string, dnsNS map[string]string) {
	// Generate zone SOA record
	re.newRecord(zone, "@ SOA ns noc-srv.lastmile.sk. "+time.Now().Format("2006010215")+" 7200 3600 1209600 3600")

	// Generate NS record for zone
	for k, v := range dnsNS {
		re.newRecord(zone, "@ NS "+k)
		re.newRecord(zone, k+" A "+v)
	}
}
