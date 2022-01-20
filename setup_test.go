package nautobotor

import (
	"context"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/jakubjastrabik/nautobotor/ramrecords"
	"github.com/miekg/dns"
)

func Test_newNautobotor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Nautobotor
		wantErr bool
	}{
		{
			name:  "Essential test of predefined records",
			input: "nautobotor {\nwebaddress :9002\n}\n",
			want: Nautobotor{
				WebAddress: ":9002",
				RM: &ramrecords.RamRecord{
					Zones: []string{"if.lastmile.sk."},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := caddy.NewTestController("dns", tt.input)

			got, err := newNautobotor(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("newNautobotor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reposEqual(t, tt.want, got) {
				t.Errorf("newNautobotor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func reposEqual(t *testing.T, e, n Nautobotor) bool {
	if e.WebAddress != n.WebAddress {
		t.Errorf("webaddress is different. Expected %v, got %v", e, n)
		return false
	}
	for i, r := range e.RM.Zones {
		if r != n.RM.Zones[i] {
			t.Errorf("zone is different. Expected %v, got %v", r, n.RM.Zones[i])
			return false
		}
	}

	ip := map[string]string{
		"test.if.lastmile.sk.":   "192.168.1.1",
		"ans-m1.if.lastmile.sk.": "172.16.5.90",
	}

	for name, i := range ip {
		reposEqualA(t, e, n, name, i)
	}

	reposEqualSOA(t, e, n)
	reposEqualNS(t, e, n)

	return true
}

func reposEqualSOA(t *testing.T, e, n Nautobotor) bool {
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)
	r.SetQuestion("if.lastmile.sk.", dns.TypeSOA)

	rcode, err := n.ServeDNS(context.Background(), rec, r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return false
	}
	if rcode != 0 {
		t.Errorf("Expected rcode %v, got %v", 0, rcode)
		return false
	}
	if rec.Msg.Answer == nil {
		t.Errorf("no response from dns query")
		return false
	}
	soa := rec.Msg.Answer[0].(*dns.SOA).Ns

	if soa != "ns.if.lastmile.sk." {
		t.Errorf("Expected %v, got %v", "ns.if.lastmile.sk.", soa)
		return false
	}

	return true
}

func reposEqualNS(t *testing.T, e, n Nautobotor) bool {
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)
	r.SetQuestion("if.lastmile.sk.", dns.TypeNS)

	rcode, err := n.ServeDNS(context.Background(), rec, r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return false
	}
	if rcode != 0 {
		t.Errorf("Expected rcode %v, got %v", 0, rcode)
		return false
	}
	if rec.Msg.Answer == nil {
		t.Errorf("no response from dns query")
		return false
	}

	ns := rec.Msg.Answer[0].(*dns.NS).String()
	if ns != "if.lastmile.sk.	3600	IN	NS	ans-m1.if.lastmile.sk." {
		t.Errorf("Expected %v, got %v", "if.lastmile.sk.	3600	IN	NS	ans-m1.if.lastmile.sk.", ns)
		return false
	}

	return true
}

func reposEqualA(t *testing.T, e, n Nautobotor, name, ip string) bool {
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)
	r.SetQuestion(name, dns.TypeA)

	rcode, err := n.ServeDNS(context.Background(), rec, r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return false
	}
	if rcode != 0 {
		t.Errorf("Expected rcode %v, got %v", 0, rcode)
		return false
	}
	if rec.Msg.Answer == nil {
		t.Errorf("no response from dns query")
		return false
	}

	ns := rec.Msg.Answer[0].(*dns.A).A.String()
	if ns != ip {
		t.Errorf("Expected %v, got %v", ip, ns)
		return false
	}

	return true
}
