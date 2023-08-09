package entitymanager

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openconfig/bootz/proto/bootz"
	"github.com/openconfig/bootz/server/service"
	"google.golang.org/protobuf/proto"
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

func TestSign(t *testing.T) {
	tests := []struct {
		desc    string
		resp    *bootz.GetBootstrapDataResponse
		wantErr bool
	}{
		{
			desc: "Success",
			resp: &bootz.GetBootstrapDataResponse{
				SignedResponse: &bootz.BootstrapDataSigned{
					Responses: []*bootz.BootstrapDataResponse{
						{SerialNum: "123A"},
					},
				},
			},
			wantErr: false,
		},
		{
			desc:    "Empty response",
			resp:    &bootz.GetBootstrapDataResponse{},
			wantErr: true,
		},
	}

	em := New()
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := em.Sign(test.resp, priv)
			if err != nil {
				if test.wantErr {
					t.Skip()
				}
				t.Errorf("Sign() err = %v, want %v", err, test.wantErr)
			}
			signedResponseBytes, err := proto.Marshal(test.resp.GetSignedResponse())
			if err != nil {
				t.Fatal(err)
			}
			hashed := sha256.Sum256(signedResponseBytes)
			err = rsa.VerifyPKCS1v15(&priv.PublicKey, crypto.SHA256, hashed[:], []byte(test.resp.GetResponseSignature()))
			if err != nil {
				t.Errorf("Sign() err == %v, want %v", err, test.wantErr)
			}
		})
	}
}
