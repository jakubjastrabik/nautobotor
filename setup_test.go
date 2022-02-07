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

	tests := []struct {
		name    string
		input   string
		want    Nautobotor
		ipAdd   []nautobot.IPaddress
		wantErr bool
	}{
		{
			name:  "Creating Record via webhook",
			input: "nautobotor {\nwebaddress :9002\nnautoboturl  http://geriatrix.if.lastmile.sk/api/ipam/ip-addresses \ntoken d4c7513f5ab6a3d42a11ed579bd7cc16acdd4b05\n}\n",
			want: Nautobotor{
				WebAddress: ":9002",
				RM: &ramrecords.RamRecord{
					Zones: []string{"if.lastmile.sk."},
				},
			},
			ipAdd: []nautobot.IPaddress{
				{
					Event: "created",
					Data: nautobot.Data{
						Address:  "172.16.5.3/24",
						Dns_name: "test.if.lastmile.sk",
						Family: nautobot.Family{
							Value: 4,
						},
						Status: nautobot.Status{
							Value: "active",
						},
					},
				},
				{
					Event: "updated",
					Data: nautobot.Data{
						Address:  "172.16.5.76/24",
						Dns_name: "sk-f1.if.lastmile.sk.",
						Family: nautobot.Family{
							Value: 4,
						},
						Status: nautobot.Status{
							Value: "active",
						},
					},
				},
				{
					Event: "updated",
					Data: nautobot.Data{
						Address:  "10.0.0.3/24",
						Dns_name: "pf.test.pf.",
						Family: nautobot.Family{
							Value: 4,
						},
						Status: nautobot.Status{
							Value: "active",
						},
					},
				},
				{
					Event: "deleted",
					Data: nautobot.Data{
						Address:  "172.16.5.76/24",
						Dns_name: "sk-f1.if.lastmile.sk.",
						Family: nautobot.Family{
							Value: 4,
						},
						Status: nautobot.Status{
							Value: "active",
						},
					},
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

			if err := got.getApiData(); (err != nil) != tt.wantErr {
				t.Errorf("Nautobotor.getApiData() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := got.onStartup(); (err != nil) != tt.wantErr {
				t.Errorf("Nautobotor.onStartup() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test Add records via webhook
			address := fmt.Sprintf("http://%s%s", tt.want.WebAddress, "/webhook")

			// Test webhook manipulation with records
			for _, i := range tt.ipAdd {
				jsonValue, err := json.Marshal(i)
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
			}

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

	// Some test IP address
	ip := map[string]string{
		"test.if.lastmile.sk.":   "172.16.5.3",
		"ans-m1.if.lastmile.sk.": "172.16.5.90",
		"pf.test.pf.":            "10.0.0.3",

		// Uncomment IF update is skipped
		// "arn-t1.if.lastmile.sk.": "172.16.5.76",

		// Uncomment IF delete is skipped
		//"sk-f1.if.lastmile.sk.": "172.16.5.76",
	}

	for question, i := range ip {
		testDNSQuestion(t, n, "A", question, i)
		testDNSQuestion(t, n, "PTR", question, i)
	}

	testDNSQuestion(t, n, "SOA", "if.lastmile.sk.", "ns.if.lastmile.sk.")
	testDNSQuestion(t, n, "NS", "if.lastmile.sk.", "")
	testDNSQuestion(t, n, "PTRNS", "5.16.172.in-addr.arpa.", "")

	return true
}

func testDNSQuestion(t *testing.T, n Nautobotor, recordType, question, ip string) bool {
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)

	// Set specific stuff for different DNS questions
	switch recordType {
	case "A":
		r.SetQuestion(question, dns.TypeA)
	case "SOA":
		r.SetQuestion(question, dns.TypeSOA)
	case "NS":
		r.SetQuestion(question, dns.TypeNS)
	case "PTR":
		a, _ := dns.ReverseAddr(ip)
		r.SetQuestion(a, dns.TypePTR)
	case "PTRNS":
		r.SetQuestion(question, dns.TypeNS)
	default:
		t.Errorf("Wrong dns.Type, got %s", recordType)
	}

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
		t.Errorf("No %s response from dns query", recordType)
		return false
	}

	// Set specific respon parser for different DNS questions
	switch recordType {
	case "A":
		a := rec.Msg.Answer[0].(*dns.A).A.String()
		if a != ip {
			t.Errorf("Expected %v, got %v", ip, a)
			return false
		}
	case "SOA":
		soa := rec.Msg.Answer[0].(*dns.SOA).Ns
		if soa != ip {
			t.Errorf("Expected %v, got %v", ip, soa)
			return false
		}
	case "NS":
		ns := rec.Msg.Answer[0].Header().Name
		if ns != question {
			t.Errorf("Expected %v, got %v", question, ns)
			return false
		}
	case "PTR":
		ptr := rec.Msg.Answer[0].(*dns.PTR).Ptr
		if ptr != question {
			t.Errorf("Expected %s, got %s", question, ptr)
			return false
		}
	case "PTRNS":
		nsptr := rec.Msg.Answer[0].Header().Name
		if nsptr != question {
			t.Errorf("Expected %s, got %s", question, nsptr)
			return false
		}
	default:
		t.Errorf("Wrong dns.Type, got %s", recordType)
	}

	return true
}
