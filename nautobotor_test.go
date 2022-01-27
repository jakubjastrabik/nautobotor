package nautobotor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/coredns/coredns/plugin"
	"github.com/jakubjastrabik/nautobotor/nautobot"
	"github.com/jakubjastrabik/nautobotor/ramrecords"
)

func TestNautobotor_onStartup(t *testing.T) {
	// Create Testing Data
	ip_add := &nautobot.IPaddress{
		Event: "created",
	}
	ip_add.Data.Family.Value = 4
	ip_add.Data.Address = "192.168.1.1/24"
	ip_add.Data.Status.Value = "active"
	ip_add.Data.Dns_name = "test.if.lastmile.sk."

	type fields struct {
		WebAddress string
		RM         *ramrecords.RamRecord
		ln         net.Listener
		mux        *http.ServeMux
		Next       plugin.Handler
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Test Elementary playload",
			wantErr: false,
			fields: fields{
				WebAddress: ":9002",
				RM:         &ramrecords.RamRecord{},
				ln:         nil,
				mux:        &http.ServeMux{},
				Next:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Nautobotor{
				WebAddress: tt.fields.WebAddress,
				RM:         tt.fields.RM,
				ln:         tt.fields.ln,
				mux:        tt.fields.mux,
				Next:       tt.fields.Next,
			}
			if err := n.onStartup(); (err != nil) != tt.wantErr {
				t.Errorf("Nautobotor.onStartup() error = %v, wantErr %v", err, tt.wantErr)
			}

			address := fmt.Sprintf("http://%s%s", n.WebAddress, "/webhook")

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
		})
	}
}
