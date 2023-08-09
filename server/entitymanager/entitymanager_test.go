package entitymanager

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openconfig/bootz/proto/bootz"
	"github.com/openconfig/bootz/server/service"
)

func TestResolveChassis(t *testing.T) {
	tests := []struct {
		desc    string
		input   *bootz.ChassisDescriptor
		want    *service.ChassisEntity
		wantErr bool
	}{
		{
			desc: "Fixed Form Factor Device",
			input: &bootz.ChassisDescriptor{
				SerialNumber: "456",
				Manufacturer: "Cisco",
			},
			want: &service.ChassisEntity{
				BootMode: "SecureOnly",
			},
		},
		{
			desc: "Modular Device",
			input: &bootz.ChassisDescriptor{
				ControlCards: []*bootz.ControlCard{
					{SerialNumber: "123A"},
					{SerialNumber: "123B"},
				},
			},
			want: &service.ChassisEntity{
				BootMode: "SecureOnly",
			},
		},
		{
			desc: "Chassis Not Found",
			input: &bootz.ChassisDescriptor{
				SerialNumber: "123",
				Manufacturer: "Cisco",
			},
			want:    nil,
			wantErr: true,
		},
	}

	em := New()

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := em.ResolveChassis(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("ResolveChassis(%v) err = %v, want %v", test.input, err, test.wantErr)
			}
			if !cmp.Equal(got, test.want) {
				t.Errorf("ResolveChassis(%v) got %v, want %v", test.input, got, test.want)
			}
		})
	}
}
