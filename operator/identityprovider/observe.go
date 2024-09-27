package identityprovider

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/minio/madmin-go/v3"
	"github.com/pkg/errors"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (i *identityProviderClient) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("observing resource")

	identityProvider, ok := mg.(*miniov1alpha1.IdentityProvider)
	if !ok {
		return managed.ExternalObservation{}, errNotIdentityProvider
	}

	cfgType := madmin.OpenidIDPCfg
	identityProviderList, err := i.ma.ListIDPConfig(ctx, cfgType)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot list identity providers")
	}

	identityProviderName := identityProvider.GetIdentityProviderName()
	for _, idp := range identityProviderList {
		if idp.Name == identityProviderName {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
		}
	}

	return managed.ExternalObservation{}, nil
}
