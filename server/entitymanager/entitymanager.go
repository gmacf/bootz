// Package entitymanager is an in-memory implementation of an entity manager that models an organization's inventory.
package entitymanager

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"

	"github.com/openconfig/bootz/proto/bootz"
	"github.com/openconfig/bootz/server/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type InMemoryEntityManager struct {
	// A mapping of chassis or control card serial number to ChassisEntity.
	inventory map[string]*service.ChassisEntity
}

func (m *InMemoryEntityManager) ResolveChassis(desc *bootz.ChassisDescriptor) (*service.ChassisEntity, error) {
	// Attempt to resolve a fixed form factor chassis.
	if desc.GetSerialNumber() != "" {
		c, ok := m.inventory[desc.GetSerialNumber()]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "fixed form factor chassis %v not found in inventory", desc.GetSerialNumber())
		}
		return c, nil
	}
	// Attempt to resolve control cards to chassis.
	for _, card := range desc.GetControlCards() {
		c, ok := m.inventory[card.GetSerialNumber()]
		if !ok {
			continue
		}
		return c, nil
	}
	return nil, status.Errorf(codes.NotFound, "chassis descriptor %v not found in inventory", desc)
}

func (m *InMemoryEntityManager) GetBootstrapData(*bootz.ControlCard) (*bootz.BootstrapDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Unimplemented")
}

func (m *InMemoryEntityManager) SetStatus(req *bootz.ReportStatusRequest) error {
	return status.Errorf(codes.Unimplemented, "Unimplemented")
}

// Sign unmarshals the SignedResponse bytes then generates a signature from its Ownership Certificate private key.
func (m *InMemoryEntityManager) Sign(resp *bootz.GetBootstrapDataResponse, priv *rsa.PrivateKey) error {
	if resp.GetSignedResponse() == nil {
		return status.Errorf(codes.InvalidArgument, "empty signed response")
	}
	signedResponseBytes, err := proto.Marshal(resp.GetSignedResponse())
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(signedResponseBytes)
	sig, err := rsa.SignPKCS1v15(nil, priv, crypto.SHA256, hashed[:])
	if err != nil {
		return err
	}
	resp.ResponseSignature = string(sig)
	return nil
}

func New() *InMemoryEntityManager {
	// TODO: Populate these values from a config file or command line flag.
	defaultChassis := service.ChassisEntity{
		BootMode: "SecureOnly",
	}

	return &InMemoryEntityManager{
		inventory: map[string]*service.ChassisEntity{
			// Control cards 123A and 123B map to a modular chassis.
			"123A": &defaultChassis,
			"123B": &defaultChassis,
			// Fixed Form Chassis 456 maps to the FFF chassis.
			"456": &defaultChassis,
		},
	}
}
