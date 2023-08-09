// Package entitymanager is an in-memory implementation of an entity manager that models an organization's inventory.
package entitymanager

import (
	"github.com/openconfig/bootz/proto/bootz"
	"github.com/openconfig/bootz/server/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (m *InMemoryEntityManager) SetStatus(*bootz.ReportStatusRequest) error {
	return status.Errorf(codes.Unimplemented, "Unimplemented")
}

func (m *InMemoryEntityManager) Sign(*bootz.GetBootstrapDataResponse) error {
	return status.Errorf(codes.Unimplemented, "Unimplemented")
}

func New() *InMemoryEntityManager {
	// TODO: Populate these values from a config file or command line flag.
	// This represents a fixed form factor chassis (e.g. no control cards).
	fffChassis := service.ChassisEntity{
		BootMode: "SecureOnly",
	}
	// This represents a modular chassis (e.g. contrains control cards).
	modularChassis := service.ChassisEntity{
		BootMode: "SecureOnly",
	}

	return &InMemoryEntityManager{
		inventory: map[string]*service.ChassisEntity{
			// Control cards 123A and 123B map to the modular chassis.
			"123A": &modularChassis,
			"123B": &modularChassis,
			// Chassis 456 maps to the FFF chassis.
			"456": &fffChassis,
		},
	}
}
