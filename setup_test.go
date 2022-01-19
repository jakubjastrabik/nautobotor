package nautobotor

import (
	"context"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin"
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
			name:  "Esential test",
			input: "nautobotor {\nwebaddress :9002\n}\n",
			want: Nautobotor{
				WebAddress: ":9002",
				RM: &ramrecords.RamRecord{
					Zones: []string{"if.lastmile.sk."},
				},
				Next: plugin.Handler(nil),
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
		return false
	}
	for i, r := range e.RM.Zones {
		if r != n.RM.Zones[i] {
			return false
		}
	}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)
	r.SetQuestion("my_host.if.lastmile.sk.", dns.TypeA)

	rcode, err := n.ServeDNS(context.Background(), rec, r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rcode != 0 {
		t.Errorf("Expected rcode %v, got %v", 0, rcode)
	}
	IP := rec.Msg.Answer[0].(*dns.A).A.String()

	if IP != "10.0.0.2" {
		t.Errorf("Expected %v, got %v", "10.0.0.2", IP)
	}

	return true
}
