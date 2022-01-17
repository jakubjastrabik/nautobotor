package nautobotor

import (
	"reflect"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin"
)

func Test_parseNawtobotor(t *testing.T) {
	type args struct {
		c *caddy.Controller
	}
	tests := []struct {
		name    string
		input   string
		args    args
		want    *Nautobotor
		wantErr bool
	}{
		{
			name:  "minimal valid config",
			input: "debug\n nautobotor {\nwebaddress :9002\n}\n",
			args:  args{},
			want: &Nautobotor{
				WebAddress: ":9002",
				RM:         nil,
				// RM:         &ramrecords.RamRecord{},
				Next: plugin.Handler(nil),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.args.c = caddy.NewTestController("dns", tt.input)
		if err := setup(tt.args.c); err != nil {
			t.Fatalf("Expected no errors, but got: %v", err)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNawtobotor(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNawtobotor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseNawtobotor() = %v, want %v", got, tt.want)
			}
		})
	}
}
