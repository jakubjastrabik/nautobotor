package nautobotor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/jakubjastrabik/nautobotor/nautobot"
	"github.com/jakubjastrabik/nautobotor/ramrecords"
	"github.com/miekg/dns"
)

func Test_newNautobotor(t *testing.T) {
	// Create Testing Data
	ip_add := &nautobot.IPaddress{
		Event: "created",
	}
	ip_add.Data.Family.Value = 4
	ip_add.Data.Address = "172.16.5.3/24"
	ip_add.Data.Status.Value = "active"
	ip_add.Data.Dns_name = "test.if.lastmile.sk"
	ip_addEdit := &nautobot.IPaddress{
		Event: "updated",
	}
	ip_addEdit.Data.Family.Value = 4
	ip_addEdit.Data.Address = "172.16.5.76/24"
	ip_addEdit.Data.Status.Value = "active"
	ip_addEdit.Data.Dns_name = "arn-f1.if.lastmile.sk"
	// ip_addDel := &nautobot.IPaddress{
	// 	Event: "deleted",
	// }
	// ip_addDel.Data.Family.Value = 4
	// ip_addDel.Data.Address = "172.16.5.76/24"
	// ip_addDel.Data.Status.Value = "active"
	// ip_addDel.Data.Dns_name = "arn-f1.if.lastmile.sk"

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

			if err := got.onStartup(); (err != nil) != tt.wantErr {
				t.Errorf("Nautobotor.onStartup() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test Add records via webhook
			address := fmt.Sprintf("http://%s%s", tt.want.WebAddress, "/webhook")
			jsonValue, err := json.Marshal(ip_add)
			if err != nil {
				t.Errorf("Nautobotor.Marshal unable correct return json data error = %s", err)
			}
			resp, err := http.Post(address, "application/json", bytes.NewBuffer(jsonValue))
			if err != nil {
				t.Errorf("Error posting JSON request error = %s", err)
			}
			if resp.StatusCode != 200 {
				t.Errorf("Error posting JSON response error = %s", resp.Status)
			}

			jsonValue, err = json.Marshal(ip_addEdit)
			if err != nil {
				t.Errorf("Nautobotor.Marshal unable correct return json data error = %s", err)
			}
			resp, err = http.Post(address, "application/json", bytes.NewBuffer(jsonValue))
			if err != nil {
				t.Errorf("Error posting JSON request error = %s", err)
			}
			if resp.StatusCode != 200 {
				t.Errorf("Error posting JSON response error = %s", resp.Status)
			}

			// jsonValue, err = json.Marshal(ip_addDel)
			// if err != nil {
			// 	t.Errorf("Nautobotor.Marshal unable correct return json data error = %s", err)
			// }
			// resp, err = http.Post(address, "application/json", bytes.NewBuffer(jsonValue))
			// if err != nil {
			// 	t.Errorf("Error posting JSON request error = %s", err)
			// }
			// if resp.StatusCode != 200 {
			// 	t.Errorf("Error posting JSON response error = %s", resp.Status)
			// }

			// test DNS response
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
		"test.if.lastmile.sk.":   "172.16.5.3",
		"ans-m1.if.lastmile.sk.": "172.16.5.90",
		// "arn-t1.if.lastmile.sk.":  "172.16.5.76",
		"arn-f1.if.lastmile.sk.": "172.16.5.76",
	}

	for name, i := range ip {
		reposEqualA(t, e, n, name, i)
	}

	reposEqualSOA(t, e, n)
	reposEqualNS(t, e, n)
	reposEqualPTR(t, e, n, "ans-m1.if.lastmile.sk.", "172.16.5.90")

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
		t.Errorf("no SOA response from dns query")
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
		t.Errorf("no NS response from dns query")
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
		t.Errorf("no A response from dns query")
		return false
	}

	ns := rec.Msg.Answer[0].(*dns.A).A.String()
	if ns != ip {
		t.Errorf("Expected %v, got %v", ip, ns)
		return false
	}

	return true
}

func reposEqualPTR(t *testing.T, e, n Nautobotor, name, ip string) bool {
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)
	a, _ := dns.ReverseAddr(ip)
	r.SetQuestion(a, dns.TypePTR)

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
		t.Errorf("no A response from dns query")
		return false
	}

	ns := rec.Msg.Answer[0].(*dns.PTR).Ptr
	if ns != name {
		t.Errorf("Expected %s, got %s", name, ns)
		return false
	}

	return true
}
