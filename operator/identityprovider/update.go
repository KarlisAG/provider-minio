package identityprovider

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (i *identityProviderClient) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("updating resource")

	_, ok := mg.(*miniov1alpha1.IdentityProvider)
	if !ok {
		return managed.ExternalUpdate{}, errNotIdentityProvider
	}

	return managed.ExternalUpdate{}, nil
}
