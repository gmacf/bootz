package entitymanager

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"

	"github.com/openconfig/bootz/proto/bootz"
	"github.com/openconfig/bootz/server/service"
	"google.golang.org/protobuf/proto"
)

type EM struct {
}

// GetBootstrapData returns the intended image, boot config and other artifacts for the given control card.
func (e *EM) GetBootstrapData(card *bootz.ControlCard) (*bootz.BootstrapDataResponse, error) {
	return nil, nil
}

// SetStatus updates the internal status of a Bootz request for an entity.
func (e *EM) SetStatus(req *bootz.ReportStatusRequest) error {
	// Get device information from metadata
	// Iterate over control cards and set the bootstrap status for element
	return nil
}

// Sign unmarshals the SignedResponse bytes then generates a signature from its Ownership Certificate private key.
func (e *EM) Sign(resp *bootz.GetBootstrapDataResponse, priv *rsa.PrivateKey) error {
	if resp.GetSignedResponse() == nil {
		return fmt.Errorf("empty signed response")
	}
	signedResponseBytes, err := proto.Marshal(resp.GetSignedResponse())
	if err != nil {
		return fmt.Errorf("unable to marshal signed response: %v", err)
	}
	hashed := sha256.Sum256(signedResponseBytes)
	sig, err := rsa.SignPKCS1v15(nil, priv, crypto.SHA256, hashed[:])
	if err != nil {
		return fmt.Errorf("unable to sign response: %v", err)
	}
	resp.ResponseSignature = string(sig)
	return nil
}

// ResolveChassis performs an internal lookup of inventory for the chassis.
func (e *EM) ResolveChassis(entity *service.EntityLookup) (*service.ChassisEntity, error) {
	return nil, nil
}
