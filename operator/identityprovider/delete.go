package identityprovider

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/minio/madmin-go/v3"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (i *identityProviderClient) Delete(ctx context.Context, mg resource.Managed) error {
	log := controllerruntime.LoggerFrom(ctx)
	log.Info("deleting resource")

	identityProvider, ok := mg.(*miniov1alpha1.IdentityProvider)
	if !ok {
		return errNotIdentityProvider
	}

	identityProvider.SetConditions(xpv1.Deleting())

	err := i.deleteIdentityProvider(ctx, identityProvider)
	if err != nil {
		return err
	}

	i.emitDeletionEvent(identityProvider)

	return nil
}

func (i *identityProviderClient) emitDeletionEvent(identityProvider *miniov1alpha1.IdentityProvider) {
	i.recorder.Event(identityProvider, event.Event{
		Type:    event.TypeNormal,
		Reason:  "Deleted",
		Message: "Identity Provider deleted",
	})
}

func (i *identityProviderClient) deleteIdentityProvider(ctx context.Context, identityProvider *miniov1alpha1.IdentityProvider) error {
	cfgType := madmin.OpenidIDPCfg
	name := identityProvider.GetIdentityProviderName()

	restart, err := i.ma.DeleteIDPConfig(ctx, cfgType, name)
	if err != nil {
		return err
	} else if restart {
		log := controllerruntime.LoggerFrom(ctx)
		log.V(1).Info("Restarting MinIO server to apply changes")
		err = i.ma.ServiceRestart(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
