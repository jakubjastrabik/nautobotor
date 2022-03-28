package nautobotor

import (
	"testing"

	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func TestAddRecord(t *testing.T) {
	zoneTestName := "example.org."
	z := NewZone(zoneTestName)

	tests := []struct {
		r dns.RR
	}{
		{test.A("bar.example.org. 3600	IN	A 173.16.5.22")},
		{test.AAAA("e.example.org. 1800 IN AAAA 2a01:7e00::f03c:91ff:fef1:6735")},
		{test.NS("example.org.	3600	IN	NS	a.iana-servers.net.")},
		{test.NS("example.org.	3600	IN	NS	b.iana-servers.net.")},
		{test.SOA("example.org.	1800	IN	SOA	test.example.org. soc.example.org. 1459281744 14400 3600 604800 14400")},
		{test.RRSIG("example.org.	1800	IN	RRSIG	NS 8 2 1800 20160428190224 20160329190224 14460 example.org. dLNC0=")},
		{test.CNAME("wild.d.example.org. IN	CNAME	alias.example.org.")},
		{test.MX("a.example.org. IN MX 10 mx.example.org.")},
		{test.SRV("_srv._tcp.example.org. 1800 IN SRV 10 5 5223 server.example.org.")},
		{test.RRSIG("example.org.	1800	IN	RRSIG	SOA 8 2 1800 20160428190224 20160329190224 14460 example.org. dLNC0=")},
	}

	for i, tt := range tests {
		err := z.Insert(tt.r)
		if err != nil {
			t.Errorf("Test %d: unable insert record %s into zone: %s", i, tt.r, zoneTestName)
		}
	}
}

func TestZone_Remove(t *testing.T) {
	zoneTestName := "example.org."
	z := NewZone(zoneTestName)

	tests := []struct {
		r dns.RR
	}{
		{test.A("bar.example.org. 3600	IN	A 173.16.5.22")},
		{test.AAAA("e.example.org. 1800 IN AAAA 2a01:7e00::f03c:91ff:fef1:6735")},
		{test.NS("example.org.	3600	IN	NS	a.iana-servers.net.")},
		{test.NS("example.org.	3600	IN	NS	b.iana-servers.net.")},
		{test.SOA("example.org.	1800	IN	SOA	test.example.org. soc.example.org. 1459281744 14400 3600 604800 14400")},
		{test.RRSIG("example.org.	1800	IN	RRSIG	NS 8 2 1800 20160428190224 20160329190224 14460 example.org. dLNC0=")},
		{test.CNAME("wild.d.example.org. IN	CNAME	alias.example.org.")},
		{test.MX("a.example.org. IN MX 10 mx.example.org.")},
		{test.SRV("_srv._tcp.example.org. 1800 IN SRV 10 5 5223 server.example.org.")},
		{test.RRSIG("example.org.	1800	IN	RRSIG	SOA 8 2 1800 20160428190224 20160329190224 14460 example.org. dLNC0=")},
	}

	delTest := []struct {
		r       dns.RR
		wantErr bool
	}{
		{test.A("bar.example.org. 3600	IN	A 173.16.5.22"), true},
	}

	for i, tt := range tests {
		err := z.Insert(tt.r)
		if err != nil {
			t.Errorf("Test %d: unable insert record %s into zone: %s", i, tt.r, zoneTestName)
		}
	}

	for i, tt := range delTest {
		err := z.Remove(tt.r)
		if err != nil {
			t.Errorf("Test %d: unable insert record %s into zone: %s", i, tt.r, zoneTestName)
		}
	}
}
