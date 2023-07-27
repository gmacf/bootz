package entitymanager

import (
	"github.com/openconfig/bootz/proto/bootz"
	"github.com/openconfig/bootz/server/service"
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
func (e *EM) Sign(resp *bootz.GetBootstrapDataResponse) error {
	return nil
}

// ResolveChassis performs an internal lookup of inventory for the chassis.
func (e *EM) ResolveChassis(entity *service.EntityLookup) (*service.ChassisEntity, error) {
	return nil, nil
}
