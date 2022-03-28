package nautobotor

import (
	"context"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func Test_newNautobotor(t *testing.T) {
	input := "nautobotor {\nwebaddress :9002\nnautoboturl  http://geriatrix.test.org./api/ipam/ip-addresses \ntoken d4c7513f5ab6a3d42a11ed579bd7cc16acdd4b05\n}\n"

	c := caddy.NewTestController("dns", input)

	got, err := newNautobotor(c)
	if err != nil {
		t.Errorf("newNautobotor() error = %v", err)
		return
	}

	if err := got.getApiData(); err != nil {
		t.Errorf("Nautobotor.getApiData() error = %v", err)
	}

	if err := got.onStartup(); err != nil {
		t.Errorf("Nautobotor.onStartup() error = %v", err)
	}

	tests := []struct {
		zoneName string
		rr       []dns.RR
	}{
		{
			zoneName: "example.org.",
			rr: []dns.RR{
				test.A("bar.example.org. 3600	IN	A 173.16.5.22"),
				test.A("bag.example.org. 3600	IN	A 173.16.5.23"),
				test.A("bat.example.org. 3600	IN	A 173.16.5.24"),
				test.NS("example.org.	3600	IN	NS	a.iana-servers.net."),
				test.NS("example.org.	3600	IN	NS	b.iana-servers.net."),
				test.SOA("example.org.	1800	IN	SOA	test.example.org. soc.example.org. 1459281744 14400 3600 604800 14400"),
			},
		},
		{
			zoneName: "test.org.",
			rr: []dns.RR{
				test.A("bar.test.org. 3600	IN	A 173.16.10.22"),
				test.A("bag.test.org. 3600	IN	A 173.16.10.23"),
				test.A("bat.test.org. 3600	IN	A 173.16.10.24"),
				test.NS("test.org.	3600	IN	NS	a.iana-servers.net."),
				test.NS("test.org.	3600	IN	NS	b.iana-servers.net."),
				test.SOA("test.org.	1800	IN	SOA	test.test.org. soc.test.org. 1459281744 14400 3600 604800 14400"),
			},
		},
	}
	testsDel := []struct {
		zoneName string
		rr       []dns.RR
	}{
		{
			zoneName: "example.org.",
			rr:       []dns.RR{
				// test.A("bar.example.org. 3600	IN	A 173.16.5.22"),
				// test.A("bat.example.org. 3600	IN	A 173.16.5.24"),
			},
		},
	}

	testsResp := []struct {
		proto  string
		domain string
		record string
	}{
		{"SOA", "test.org.", "test.test.org."},
		{"SOA", "example.org.", "test.example.org."},
		{"A", "bat.test.org.", "173.16.10.24"},

		{"A", "bat.example.org.", "173.16.5.24"},
		{"A", "bar.example.org.", "173.16.5.22"},

		{"PTRNS", "5.16.172.in-addr.arpa.", ""},

		{"NS", "example.org.", ""},
		{"NS", "test.org.", ""},
	}

	for i := range tests {
		// Add zone from test
		got.Zones.AddZone(tests[i].zoneName, "")
		for _, r := range tests[i].rr {
			got.Zones.Z[tests[i].zoneName].Insert(r)
		}
	}

	for i := range testsDel {
		for _, r := range testsDel[i].rr {
			got.Zones.Z[testsDel[i].zoneName].Remove(r)
		}
	}

	// Test DNS record via lookup
	for _, tr := range testsResp {
		testDNSQuestion(t, got, tr.proto, tr.domain, tr.record)
	}

}

// func Test_webHook(t *testing.T) {

// 	tests := []struct {
// 		name    string
// 		input   string
// 		want    Nautobotor
// 		ipAdd   []nautobot.IPaddress
// 		wantErr bool
// 	}{
// 		{
// 			name:  "Creating Record via webhook",
// 			input: "nautobotor {\nwebaddress :9002\nnautoboturl  http://geriatrix.test.org./api/ipam/ip-addresses \ntoken d4c7513f5ab6a3d42a11ed579bd7cc16acdd4b05\n}\n",
// 			want: Nautobotor{
// 				WebAddress: ":9002",
// 			},
// 			ipAdd: []nautobot.IPaddress{
// 				{
// 					Event: "created",
// 					Data: nautobot.Data{
// 						Address:  "10.1.1.4/24",
// 						Dns_name: "test.cc.example.org",
// 						Family: nautobot.Family{
// 							Label: "IPv4",
// 						},
// 						Status: nautobot.Status{
// 							Value: "active",
// 						},
// 					},
// 				},
// 				// {
// 				// 	Event: "updated",
// 				// 	Data: nautobot.Data{
// 				// 		Address:  "10.5.1.4/24",
// 				// 		Dns_name: "sk-f1.cc.example.org",
// 				// 		Family: nautobot.Family{
// 				// 			Label: "IPv4",
// 				// 		},
// 				// 		Status: nautobot.Status{
// 				// 			Value: "active",
// 				// 		},
// 				// 	},
// 				// },
// 				// {
// 				// 	Event: "updated",
// 				// 	Data: nautobot.Data{
// 				// 		Address:  "10.0.0.3/24",
// 				// 		Dns_name: "pf.test.pf",
// 				// 		Family: nautobot.Family{
// 				// 			Label: "IPv4",
// 				// 		},
// 				// 		Status: nautobot.Status{
// 				// 			Value: "active",
// 				// 		},
// 				// 	},
// 				// },
// 				{
// 					Event: "deleted",
// 					Data: nautobot.Data{
// 						Address:  "10.1.1.4/24",
// 						Dns_name: "test.cc.example.org",
// 						Family: nautobot.Family{
// 							Label: "IPv4",
// 						},
// 						Status: nautobot.Status{
// 							Value: "active",
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := caddy.NewTestController("dns", tt.input)

// 			got, err := newNautobotor(c)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("newNautobotor() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if err := got.getApiData(); (err != nil) != tt.wantErr {
// 				t.Errorf("Nautobotor.getApiData() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			if err := got.onStartup(); (err != nil) != tt.wantErr {
// 				t.Errorf("Nautobotor.onStartup() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			// Test Add records via webhook
// 			address := fmt.Sprintf("http://%s%s", tt.want.WebAddress, "/webhook")

// 			// Test webhook manipulation with records
// 			for _, i := range tt.ipAdd {
// 				jsonValue, err := json.Marshal(i)
// 				if err != nil {
// 					t.Errorf("Nautobotor.Marshal unable correct return json data error = %s", err)
// 				}
// 				resp, err := http.Post(address, "application/json", bytes.NewBuffer(jsonValue))
// 				if err != nil {
// 					t.Errorf("Error posting JSON request error = %s", err)
// 				}
// 				if resp.StatusCode != 200 {
// 					t.Errorf("Error posting JSON response error = %s", resp.Status)
// 				}
// 			}

// 			// test DNS response
// 			if !reposEqual(t, tt.want, got) {
// 				t.Errorf("newNautobotor() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func reposEqual(t *testing.T, e, n Nautobotor) bool {
// 	if e.WebAddress != n.WebAddress {
// 		t.Errorf("webaddress is different. Expected %v, got %v", e, n)
// 		return false
// 	}

// 	// Some test IP address
// 	ip := map[string]string{
// 		// "test.cc.example.org.": "10.1.1.4",
// 		// "pf.test.pf.":            "10.0.0.3",

// 		// Uncomment IF update is skipped
// 		// "arn-t1.test.org..": "10.5.1.4",

// 		// Uncomment IF delete is skipped
// 		//"sk-f1.test.org..": "10.5.1.4",
// 	}

// 	for question, i := range ip {
// 		testDNSQuestion(t, n, "A", question, i)
// 		testDNSQuestion(t, n, "PTR", question, i)
// 	}

// 	testDNSQuestion(t, n, "SOA", "test.org..", "ns.test.org..")
// 	testDNSQuestion(t, n, "NS", "test.org..", "")
// 	testDNSQuestion(t, n, "PTRNS", "5.16.172.in-addr.arpa.", "")

// 	return true
// }

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
		t.Error(rec.Msg)
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
