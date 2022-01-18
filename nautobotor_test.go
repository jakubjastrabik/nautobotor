package nautobotor

import (
	"reflect"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/test"
	"github.com/jakubjastrabik/nautobotor/nautobot"
	"github.com/jakubjastrabik/nautobotor/ramrecords"
	"github.com/miekg/dns"
)

func TestNautobotor_handleData(t *testing.T) {

	tSOA := test.SOA("if.lastmile.sk.	60	IN	SOA	ns.if.lastmile.sk. noc-srv.lastmile.sk. " + time.Now().Format("2006010215") + " 7200 3600 1209600 3600")
	tNS1 := test.NS("if.lastmile.sk.	60	IN	NS	arn-x1.if.lastmile.sk.")
	tNS2 := test.NS("if.lastmile.sk.	60	IN	NS	ans-m1.if.lastmile.sk.")
	tNS3 := test.NS("if.lastmile.sk.	60	IN	NS	arn-t1.if.lastmile.sk.")
	tAN1 := test.A("arn-x1.if.lastmile.sk.	60	IN	A	172.16.5.77")
	tAN2 := test.A("ans-m1.if.lastmile.sk.	60	IN	A	172.16.5.90")
	tAN3 := test.A("arn-t1.if.lastmile.sk.	60	IN	A	172.16.5.76")
	tA1 := test.A("test.if.lastmile.sk.	60	IN	A	192.168.0.1")

	type args struct {
		c *caddy.Controller
	}
	tests := []struct {
		name    string
		input   string
		args    args
		want    *Nautobotor
		ip      *nautobot.IPaddress
		wantErr bool
	}{
		{
			name:  "minimal valid config",
			input: "debug\n nautobotor {\nwebaddress :9002\n}\n",
			args:  args{},
			want: &Nautobotor{
				WebAddress: ":9002",
				// RM:         nil,
				RM: &ramrecords.RamRecord{
					Zones: []string{"if.lastmile.sk"},
					M:     map[string][]dns.RR{"if.lastmile.sk": {tSOA, tNS1, tNS2, tNS3, tAN1, tAN2, tAN3, tA1}},
				},
				Next: plugin.Handler(nil),
			},
			ip: &nautobot.IPaddress{
				Event: "created",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.ip.Data.Family.Value = 4
		tt.ip.Data.Address = "192.168.0.1/24"
		tt.ip.Data.Status.Value = "active"
		tt.ip.Data.Dns_name = "test.if.lastmile.sk"

		tt.args.c = caddy.NewTestController("dns", tt.input)

		if err := setup(tt.args.c); err != nil {
			t.Fatalf("Expected no errors, but got: %v", err)
		}

		t.Run(tt.name, func(t *testing.T) {
			n, err := parseNawtobotor(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNawtobotor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			n.RM, err = ramrecords.NewRamRecords()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRamRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := n.handleData(tt.ip); (err != nil) != tt.wantErr {
				t.Errorf("Nautobotor.handleData() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(n, tt.want) {
				t.Errorf("parseNawtobotor() = %v, want %v", n, tt.want)
			}

		})
	}
}
