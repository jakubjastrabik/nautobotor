package nautobotor

import "testing"

func TestZones_AddZone(t *testing.T) {
	type fields struct {
		Z     map[string]*Zone
		Names []string
	}
	type args struct {
		name string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Basic test",
			fields: fields{
				Z:     map[string]*Zone{},
				Names: []string{"example.org."},
			},
			args:    args{"example.org."},
			wantErr: false,
		},
		{
			name: "Basic Error test",
			fields: fields{
				Z:     map[string]*Zone{},
				Names: []string{"example.org."},
			},
			args:    args{"test.org."},
			wantErr: true,
		},
	}
	z := Zones{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := z.AddZone(tt.args.name); err != nil {
				t.Errorf("Zones.AddZone() error = %v", err)
			}
			for _, name := range tt.fields.Names {
				if _, ok := z.Z[name]; !ok {
					if !tt.wantErr {
						t.Errorf("Zones.AddZone() Zone does not exist for name %v", name)
					}
				}
			}
		})
	}
}
