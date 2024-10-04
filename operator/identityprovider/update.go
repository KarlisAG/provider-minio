package identityprovider

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	miniov1 "github.com/vshn/provider-minio/apis/minio/v1"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (i *identityProviderClient) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("updating resource")

	identityProvider, ok := mg.(*miniov1alpha1.IdentityProvider)
	if !ok {
		return managed.ExternalUpdate{}, errNotIdentityProvider
	}

	identityProvider.SetConditions(miniov1.Updating())

	err := i.updateIdentityProvider(ctx, identityProvider)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	i.emitUpdateEvent(identityProvider)

	return managed.ExternalUpdate{}, nil
}

func (i *identityProviderClient) updateIdentityProvider(ctx context.Context, identityProvider *miniov1alpha1.IdentityProvider) error {
	return i.createOrUpdateIdentityProvider(ctx, identityProvider, true)
}

func (i *identityProviderClient) emitUpdateEvent(identityProvider *miniov1alpha1.IdentityProvider) {
	i.recorder.Event(identityProvider, event.Event{
		Type:    event.TypeNormal,
		Reason:  "Updated",
		Message: "Identity Provider successfully updated",
	})
}
