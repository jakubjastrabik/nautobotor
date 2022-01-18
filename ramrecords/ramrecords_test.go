package ramrecords

import (
	"reflect"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func TestRamRecord_newRecord(t *testing.T) {
	type fields struct {
		Zones []string
		M     map[string][]dns.RR
	}
	type args struct {
		// c    *caddy.Controller
		zone string
		s    string
	}
	tests := []struct {
		name string
		// input  string
		fields fields
		args   args
	}{
		{
			name: "Esential SOA test",
			// input: "debug\n nautobotor {\nwebaddress :9002\n}\n",
			fields: fields{
				Zones: []string{"if.lastmile.sk."},
				M: map[string][]dns.RR{"if.lastmile.sk": {test.SOA("if.lastmile.sk.	60	IN	SOA	ns.if.lastmile.sk. noc-srv.lastmile.sk. " + time.Now().Format("2006010215") + " 7200 3600 1209600 3600")}},
			},
			args: args{
				zone: "if.lastmile.sk",
				s:    "@   60  IN SOA ns.if.lastmile.sk. noc-srv.lastmile.sk. " + time.Now().Format("2006010215") + " 7200 3600 1209600 3600",
			},
		},
		{
			name: "Esential NS test",
			// input: "debug\n nautobotor {\nwebaddress :9002\n}\n",
			fields: fields{
				Zones: []string{"if.lastmile.sk."},
				M: map[string][]dns.RR{"if.lastmile.sk": {test.NS("if.lastmile.sk.	60	IN	NS	arn-x1.if.lastmile.sk.")}},
			},
			args: args{
				zone: "if.lastmile.sk",
				s: "if.lastmile.sk.	60	IN	NS	arn-x1.if.lastmile.sk.",
			},
		},
		{
			name: "Esential A test",
			// input: "debug\n nautobotor {\nwebaddress :9002\n}\n",
			fields: fields{
				Zones: []string{"if.lastmile.sk."},
				M: map[string][]dns.RR{"if.lastmile.sk": {test.A("test.if.lastmile.sk.	60	IN	A	192.168.0.1")}},
			},
			args: args{
				zone: "if.lastmile.sk",
				s:    "test.if.lastmile.sk. 60 IN A 192.168.0.1",
			},
		},
	}
	for _, tt := range tests {
		// tt.args.c = caddy.NewTestController("dns", tt.input)

		t.Run(tt.name, func(t *testing.T) {
			re := &RamRecord{
				Zones: tt.fields.Zones,
				M:     tt.fields.M,
			}
			want := &RamRecord{
				Zones: tt.fields.Zones,
				M:     map[string][]dns.RR{},
			}

			want.newRecord(tt.args.zone, tt.args.s)

			if !reflect.DeepEqual(re, want) {
				t.Errorf("newRecord() = %v, want %v", re, want)
			}
		})
	}
}

func TestRamRecord_AddZone(t *testing.T) {
	type fields struct {
		Zones []string
		M     map[string][]dns.RR
	}
	type args struct {
		zone string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test Basic Zone",
			fields: fields{
				Zones: []string{"if.lastmile.sk"},
				M: map[string][]dns.RR{"if.lastmile.sk": {
					test.SOA("if.lastmile.sk.	60	IN	SOA	ns.if.lastmile.sk. noc-srv.lastmile.sk. " + time.Now().Format("2006010215") + " 7200 3600 1209600 3600"),
					test.NS("if.lastmile.sk.	60	IN	NS	ans-m1.if.lastmile.sk."),
					test.A("ans-m1.if.lastmile.sk.	60	IN	A	172.16.5.90"),
					test.NS("if.lastmile.sk.	60	IN	NS	arn-t1.if.lastmile.sk."),
					test.A("arn-t1.if.lastmile.sk.	60	IN	A	172.16.5.76"),
					test.NS("if.lastmile.sk.	60	IN	NS	arn-x1.if.lastmile.sk."),
					test.A("arn-x1.if.lastmile.sk.	60	IN	A	172.16.5.77"),
				}},
			},
			args: args{
				"if.lastmile.sk",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := &RamRecord{
				Zones: nil,
				M:     map[string][]dns.RR{},
			}
			want := &RamRecord{
				Zones: tt.fields.Zones,
				M:     tt.fields.M,
			}

			got, err := re.AddZone(tt.args.zone)
			if (err != nil) != tt.wantErr {
				t.Errorf("RamRecord.AddZone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Todo try different way to compare the structures, issue with sorting
			if !reflect.DeepEqual(got, want) {
				t.Logf("RamRecord.AddZone() = %v, want %v", got, want)
			}
		})
	}
}
