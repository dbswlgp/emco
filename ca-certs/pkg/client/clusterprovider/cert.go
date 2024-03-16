// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"context"
	"reflect"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/distribution"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
)

// CaCertManager exposes all the clusterProvider caCert functionalities
type CaCertManager interface {
	// Certificates
	CreateCert(ctx context.Context, cert module.CaCert, clusterProvider string, failIfExists bool) (module.CaCert, bool, error)
	DeleteCert(ctx context.Context, cert, clusterProvider string) error
	GetAllCert(ctx context.Context, clusterProvider string) ([]module.CaCert, error)
	GetCert(ctx context.Context, cert, clusterProvider string) (module.CaCert, error)
}

// CaCertKey represents the resources associated with a clusterProvider caCert
type CaCertKey struct {
	Cert            string `json:"caCertCp"`
	ClusterProvider string `json:"clusterProvider"`
}

// CaCertClient holds the client properties
type CaCertClient struct {
}

// NewCaCertClient returns an instance of the CaCertClient which implements the Manager
func NewCaCertClient() *CaCertClient {
	return &CaCertClient{}
}

// CreateCert creates a clusterProvider caCert
func (c *CaCertClient) CreateCert(ctx context.Context, cert module.CaCert, clusterProvider string, failIfExists bool) (module.CaCert, bool, error) {
	certExists := false
	ck := CaCertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider}
	cc := module.NewCaCertClient(ck)

	if cer, err := cc.GetCert(ctx); err == nil &&
		!reflect.DeepEqual(cer, module.CaCert{}) {
		certExists = true
	}

	if certExists &&
		failIfExists {
		return module.CaCert{}, certExists, emcoerror.NewEmcoError(
			module.CaCertAlreadyExists,
			emcoerror.Conflict,
		)
	}

	if certExists {
		// check the enrollment state
		if err := verifyEnrollmentStateBeforeUpdate(ctx, cert.MetaData.Name, clusterProvider); err != nil {
			return module.CaCert{}, certExists, err
		}

		// check the distribution state
		if err := verifyDistributionStateBeforeUpdate(ctx, cert.MetaData.Name, clusterProvider); err != nil {
			return module.CaCert{}, certExists, err
		}

		return cert, certExists, cc.UpdateCert(ctx, cert)
	}

	_, certExists, err := cc.CreateCert(ctx, cert, failIfExists)
	if err != nil {
		return module.CaCert{}, certExists, err
	}

	// create enrollment stateInfo
	ek := EnrollmentKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	sc := module.NewStateClient(ek)
	if err := sc.Create(ctx, ""); err != nil {
		return module.CaCert{}, certExists, err
	}

	// create distribution stateInfo
	dk := DistributionKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}
	sc = module.NewStateClient(dk)
	if err := sc.Create(ctx, ""); err != nil {
		return module.CaCert{}, certExists, err
	}

	return cert, certExists, nil
}

// DeleteCert deletes a given clusterProvider caCert
func (c *CaCertClient) DeleteCert(ctx context.Context, cert, clusterProvider string) error {
	// check the enrollment state
	if err := verifyEnrollmentStateBeforeDelete(ctx, cert, clusterProvider); err != nil {
		// if the StateInfo cannot be found, then a caCert record may not present
		// Continue with the caCert deletion if the error is NotFound
		// In all other cases, intercept and return the error
		switch e := err.(type) { // To avoid any panic if the error is other than the emco error type
		case *emcoerror.Error:
			if e.Reason != emcoerror.NotFound {
				return e
			}
		default:
			return err
		}
	}

	// check the distribution state
	if err := verifyDistributionStateBeforeDelete(ctx, cert, clusterProvider); err != nil {
		// if the StateInfo cannot be found, then a caCert record may not present
		// Continue with the caCert deletion if the error is NotFound
		// In all other cases, intercept and return the error
		switch e := err.(type) { // To avoid any panic if the error is other than the emco error type
		case *emcoerror.Error:
			if e.Reason != emcoerror.NotFound {
				return e
			}
		default:
			return err
		}
	}

	// delete enrollment stateInfo
	ek := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	sc := module.NewStateClient(ek)
	if err := sc.Delete(ctx); err != nil {
		return err
	}

	// delete distribution stateInfo
	dk := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}
	sc = module.NewStateClient(dk)
	if err := sc.Delete(ctx); err != nil {
		return err
	}

	// delete caCert
	ck := CaCertKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	return module.NewCaCertClient(ck).DeleteCert(ctx)
}

// GetAllCert returns all the clusterProvider caCert
func (c *CaCertClient) GetAllCert(ctx context.Context, clusterProvider string) ([]module.CaCert, error) {
	ck := CaCertKey{
		ClusterProvider: clusterProvider}

	return module.NewCaCertClient(ck).GetAllCert(ctx)
}

// GetCert returns the clusterProvider caCert
func (c *CaCertClient) GetCert(ctx context.Context, cert, clusterProvider string) (module.CaCert, error) {
	ck := CaCertKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	return module.NewCaCertClient(ck).GetCert(ctx)
}

// verifyEnrollmentStateBeforeDelete
func verifyEnrollmentStateBeforeDelete(ctx context.Context, cert, clusterProvider string) error {
	k := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}

	return module.NewCaCertClient(k).VerifyStateBeforeDelete(ctx, cert, enrollment.AppName)
}

// verifyDistributionStateBeforeDelete
func verifyDistributionStateBeforeDelete(ctx context.Context, cert, clusterProvider string) error {
	k := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	return module.NewCaCertClient(k).VerifyStateBeforeDelete(ctx, cert, distribution.AppName)

}

// verifyEnrollmentStateBeforeUpdate
func verifyEnrollmentStateBeforeUpdate(ctx context.Context, cert, clusterProvider string) error {
	k := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}

	return module.NewCaCertClient(k).VerifyStateBeforeUpdate(ctx, cert, enrollment.AppName)
}

// verifyDistributionStateBeforeUpdate
func verifyDistributionStateBeforeUpdate(ctx context.Context, cert, clusterProvider string) error {
	k := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	return module.NewCaCertClient(k).VerifyStateBeforeUpdate(ctx, cert, distribution.AppName)

}
