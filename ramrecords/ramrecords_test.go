package ramrecords

import (
	"reflect"
	"testing"
)

func TestNewRamRecords(t *testing.T) {
	tests := []struct {
		name    string
		want    *RamRecord
		wantErr bool
	}{
		// {name: "Test", wantErr: false, want: &RamRecord{Zones: []string{"if.lastmile.sk."}, M: map[string][]dns.RR{"if.lastmile.sk.": {test.A("test.if.lastmile.sk 60 A 192.168.1.1")}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRamRecords()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRamRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRamRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRamRecords(t *testing.T) {
	re, _ := NewRamRecords()
	_, err := re.AddZone("if.lastmile.sk")
	if err != nil {
		t.Error(err)
	}
	ss, _ := re.AddRecord(4, "192.168.1.251/24", "test.if.lastmile.sk", "if.lastmile.sk")
	t.Log(ss)

}
