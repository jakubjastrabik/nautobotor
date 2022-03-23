package nautobotor

import (
	"reflect"
	"testing"

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createRRString(tt.args.t, tt.args.fqdn, tt.args.ip); got != tt.want {
				t.Errorf("createRRString() = %v, want %v", got, tt.want)
			}
		})
	}
}
