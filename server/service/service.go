// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package service receives bootstrap requests and responds with the relevant data.
package service

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"

	"github.com/openconfig/gnmi/errlist"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	log "github.com/golang/glog"
	bpb "github.com/openconfig/bootz/proto/bootz"
)

// OVList is a mapping of control card serial number to ownership voucher.
type OVList map[string][]byte

// SecurityArtifacts contains all KeyPairs and OVs needed for the Bootz Server.
// Currently, RSA is the only encryption standard supported by these artifacts.
type SecurityArtifacts struct {
	// The Ownership Certificate is an x509 certificate/private key pair signed by the PDC.
	// The certificate is presented to the device during bootstrapping and is used to validate the Ownership Voucher.
	OwnerCert           *x509.Certificate
	OwnerCertPrivateKey crypto.PrivateKey
	// The Pinned Domain Certificate is an x509 certificate/private key pair which acts as a certificate authority on the owner's side.
	// This certificate is included in OVs.
	PDC           *x509.Certificate
	PDCPrivateKey crypto.PrivateKey
	// The Vendor CA represents a certificate authority on the vendor side. This CA signs Ownership Vouchers which are verified by the device.
	VendorCA           *x509.Certificate
	VendorCAPrivateKey crypto.PrivateKey
	// The Trust Anchor is a self signed CA used to generate the TLS certificate.
	TrustAnchor           *x509.Certificate
	TrustAnchorPrivateKey crypto.PrivateKey
	// Ownership Vouchers are a list of PKCS7 messages signed by the Vendor CA. There is one per control card.
	OV OVList
	// The TLSKeypair is a TLS certificate used to secure connections between device and server. It is derived from the Trust Anchor.
	TLSKeypair *tls.Certificate
}

// EntityManager maintains the entities and their states.
type EntityManager interface {
	ResolveChassis(*bpb.ChassisDescriptor) (*bpb.Chassis, error)
	GetBootstrapData(*bpb.Chassis, *bpb.ControlCard) (*bpb.BootstrapDataResponse, error)
	SetStatus(*bpb.ReportStatusRequest) error
	Sign(*bpb.GetBootstrapDataResponse, *bpb.Chassis, *bpb.ControlCard) error
}

// Service represents the server and entity manager.
type Service struct {
	bpb.UnimplementedBootstrapServer
	em EntityManager
}

func (s *Service) GetBootstrapData(ctx context.Context, req *bpb.GetBootstrapDataRequest) (*bpb.GetBootstrapDataResponse, error) {
	log.Infof("=============================================================================")
	log.Infof("==================== Received request for bootstrap data ====================")
	log.Infof("=============================================================================")
	fixedChasis := true
	chassisDesc := req.GetChassisDescriptor()
	log.Infof("Requesting for %v chassis %v", chassisDesc.GetManufacturer(), chassisDesc.GetSerialNumber())
	// Validate the chassis can be serviced
	chassis, err := s.em.ResolveChassis(chassisDesc)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to resolve chassis to inventory %+v, err: %v", chassisDesc, err)
	}
	log.Infof("Verified server can resolve chassis")

	// If chassis can only be booted into secure mode then return error
	if chassis.GetBootMode() == bpb.BootMode_BOOT_MODE_SECURE && req.GetNonce() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "chassis requires secure boot only")
	}

	// Iterate over the control cards and fetch data for each card.
	var errs errlist.List

	log.Infof("=============================================================================")
	log.Infof("==================== Fetching data for each control card ====================")
	log.Infof("=============================================================================")
	var responses []*bpb.BootstrapDataResponse
	var activeControlCard *bpb.ControlCard
	for _, v := range chassisDesc.GetControlCards() {
		if v.GetSerialNumber() == req.GetControlCardState().GetSerialNumber() {
			activeControlCard = v
		}
		bootdata, err := s.em.GetBootstrapData(chassis, v)
		if err != nil {
			errs.Add(err)
			log.Infof("Error occurred while retrieving data for Serial Number %v", v.GetSerialNumber())
		}
		responses = append(responses, bootdata)
	}
	if fixedChasis {
		bootdata, err := s.em.GetBootstrapData(chassis, nil)
		if err != nil {
			errs.Add(err)
			log.Infof("Error occurred while retrieving data for fixed chassis with serial number %v", chassis.GetSerialNumber())
		}
		responses = append(responses, bootdata)
	}

	if errs.Err() != nil {
		return nil, errs.Err()
	}
	log.Infof("Successfully fetched data for each control card")
	log.Infof("=============================================================================")

	nonce := req.GetNonce()
	signedResponse := &bpb.BootstrapDataSigned{
		Responses: responses,
		Nonce:     nonce,
	}
	log.Infof("Serializing the response...")
	signedResponseBytes, err := proto.Marshal(signedResponse)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully serialized the response")

	resp := &bpb.GetBootstrapDataResponse{
		// This field is deprecated but we still include it for backwards compatibility.
		SignedResponse:          signedResponse,
		SerializedBootstrapData: signedResponseBytes,
	}
	log.Infof("Response set")

	// Sign the response if Nonce is provided.
	if nonce != "" {
		log.Infof("=============================================================================")
		log.Infof("====================== Signing the response with nonce ======================")
		log.Infof("=============================================================================")
		if err := s.em.Sign(resp, chassis, activeControlCard); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to sign bootz response")
		}
		log.Infof("Signed with nonce")
	}
	log.Infof("Returning response")
	return resp, nil
}

func (s *Service) ReportStatus(ctx context.Context, req *bpb.ReportStatusRequest) (*bpb.EmptyResponse, error) {
	log.Infof("=============================================================================")
	log.Infof("========================== Status report received ===========================")
	log.Infof("=============================================================================")
	return &bpb.EmptyResponse{}, s.em.SetStatus(req)
}

// SetDeviceConfiguration is a public API for allowing the device configuration to be set for each device the
// will be responsible for configuring.  This will be only available for testing.
func (s *Service) SetDeviceConfiguration(ctx context.Context) error {
	return status.Errorf(codes.Unimplemented, "Unimplemented")
}

// New creates a new service.
func New(em EntityManager) *Service {
	return &Service{
		em: em,
	}
}
