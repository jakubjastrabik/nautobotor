package nautobotor

import (
	"reflect"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func Test_handleCreateNewRR(t *testing.T) {
	type args struct {
		zone string
		s    string
	}
	tests := []struct {
		name string
		args args
		want dns.RR
	}{
		{
			name: "Test A record",
			args: args{
				zone: "example.org.",
				s:    "test A 10.10.10.10",
			},
			want: test.A("test.example.org. 3600	IN	A 10.10.10.10"),
		},
		{
			name: "Test AAAA record",
			args: args{
				zone: "com.",
				s:    "google AAAA 2a00:1450:4014:80e::200e",
			},
			want: test.AAAA("google.com. 3600	IN	AAAA 2a00:1450:4014:80e::200e"),
		},
		{
			name: "Test NS record",
			args: args{
				zone: "example.org.",
				s:    "@ NS ns1",
			},
			want: test.NS("example.org.	3600	IN	NS	ns1.example.org."),
		},
		{
			name: "Test SOA record",
			args: args{
				zone: "example.org.",
				s:    "@ SOA ns dns-admin " + time.Now().Format("2006010215") + " 7200 3600 1209600 3600",
			},
			want: test.SOA("example.org.	3600	IN	SOA	ns.example.org. dns-admin.example.org. 2022032515 7200 3600 1209600 3600"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handleCreateNewRR(tt.args.zone, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleCreateNewRR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cutCIDRMask(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Cut Mask",
			args: args{
				ip: "10.10.10.10/32",
			},
			want: "10.10.10.10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cutCIDRMask(tt.args.ip); got != tt.want {
				t.Errorf("cutCIDRMask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createRRString(t *testing.T) {
	type args struct {
		t    string
		fqdn string
		ip   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test A",
			args: args{
				t:    "IPv4",
				fqdn: "test.example.org",
				ip:   "10.10.10.10/32",
			},
			want: "test A 10.10.10.10",
		},
		{
			name: "Test AAAA",
			args: args{
				t:    "IPv6",
				fqdn: "google.com.",
				ip:   "2a00:1450:4014:80e::200e/32",
			},
			want: "google AAAA 2a00:1450:4014:80e::200e",
		},
		{
			name: "Test NS",
			args: args{
				t:    "NS",
				fqdn: "ns1",
				ip:   "",
			},
			want: "@ NS ns1",
		},
		{
			name: "Test SOA",
			args: args{
				t:    "SOA",
				fqdn: "ns",
				ip:   "",
			},
			want: "@ SOA ns dns-admin " + time.Now().Format("2006010215") + " 7200 3600 1209600 3600",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createRRString(tt.args.t, tt.args.fqdn, tt.args.ip); got != tt.want {
				t.Errorf("createRRString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePTRzone(t *testing.T) {
	type args struct {
		ipFamily string
		ip       string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "pars IPv4 PTR zone",
			args: args{
				ipFamily: "IPv4",
				ip:       "127.10.15.1/24",
			},
			want: "15.10.127.in-addr.arpa.",
		},
		{
			name: "pars IPv6 PTR zone",
			args: args{
				ipFamily: "IPv6",
				ip:       "2a00:1450:4014:80e::200e/32",
			},
			want: "0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.5.4.1.0.0.a.2.ip6.arpa.",
		},
		{
			name: "pars IPv6 PTR zone",
			args: args{
				ipFamily: "IPv6",
				ip:       "2001:db8::/48",
			},
			want: "0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePTRzone(tt.args.ipFamily, tt.args.ip); got != tt.want {
				t.Errorf("parsePTRzone() = %v, want %v", got, tt.want)
			}
		})
	}
}
