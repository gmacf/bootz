package service

import (
	"context"
	"crypto/rsa"
	"crypto/x509"

	"github.com/openconfig/bootz/proto/bootz"
	"github.com/openconfig/gnmi/errlist"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EntityLookup struct {
	Manufacturer string
	SerialNumber string
	DeviceName   string
}

type ChassisEntity struct {
	BootMode string
}

type EntityManager interface {
	ResolveChassis(*EntityLookup) (*ChassisEntity, error)
	GetBootstrapData(*bootz.ControlCard) (*bootz.BootstrapDataResponse, error)
	SetStatus(*bootz.ReportStatusRequest) error
	Sign(*bootz.GetBootstrapDataResponse, *rsa.PrivateKey) error
}

type Service struct {
	bootz.UnimplementedBootstrapServer
	em     EntityManager
	oc     *x509.Certificate
	ocPriv *rsa.PrivateKey
}

func New(em EntityManager) *Service {
	// TODO: Populate x509 Cert and RSA Private key with real values (or generate them).
	return &Service{
		em:     em,
		oc:     &x509.Certificate{},
		ocPriv: &rsa.PrivateKey{},
	}
}

func (s *Service) GetBootstrapRequest(ctx context.Context, req *bootz.GetBootstrapDataRequest) (*bootz.GetBootstrapDataResponse, error) {
	if len(req.ChassisDescriptor.ControlCards) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "request must include at least one control card")
	}
	// Validate the chassis can be serviced.
	// TODO: Populate DeviceName.
	chassis, err := s.em.ResolveChassis(
		&EntityLookup{
			req.ChassisDescriptor.Manufacturer,
			req.ChassisDescriptor.SerialNumber,
			"",
		},
	)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to resolve chassis to inventory %+v", req.ChassisDescriptor)
	}

	// If chassis can only be booted into secure mode then return error
	if chassis.BootMode == "SecureOnly" && req.Nonce == "" {
		return nil, status.Errorf(codes.InvalidArgument, "chassis requires secure boot only")
	}

	// Iterate over the control cards and fetch data for each card.
	var errors errlist.List

	var responses []*bootz.BootstrapDataResponse
	for _, v := range req.ChassisDescriptor.ControlCards {
		data, err := s.em.GetBootstrapData(v)
		if err != nil {
			errors.Add(err)
		}
		responses = append(responses, data)
	}
	if errs := errors.Err(); errs != nil {
		return nil, errs
	}
	resp := &bootz.GetBootstrapDataResponse{
		SignedResponse: &bootz.BootstrapDataSigned{
			Responses: responses,
		},
	}
	// Sign the response if Nonce is provided.
	if req.Nonce != "" {
		if err := s.em.Sign(resp, s.ocPriv); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to sign bootz response")
		}
	}
	return resp, nil
}

func (s *Service) ReportStatus(ctx context.Context, req *bootz.ReportStatusRequest) (*bootz.EmptyResponse, error) {
	return nil, s.em.SetStatus(req)
}

// Public API for allowing the device configuration to be set for each device the
// will be responsible for configuring.  This will be only availble for testing.
// func (s *Service) SetDeviceConfiguration(ctx context.Context, req entity.ConfigurationRequest) (entity.ConfigurationResonse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "Unimplemented")
// }

func (s *Service) Start() error {
	return nil
}
