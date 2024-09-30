package identityprovider

import (
	"context"
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/minio/madmin-go/v3"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (i *identityProviderClient) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("creating resource")

	identityProvider, ok := mg.(*miniov1alpha1.IdentityProvider)
	if !ok {
		return managed.ExternalCreation{}, errNotIdentityProvider
	}

	identityProvider.SetConditions(xpv1.Creating())

	err := i.createIdentityProvider(ctx, identityProvider)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	return managed.ExternalCreation{}, i.emitCreationEvent(identityProvider)
}

func (i *identityProviderClient) emitCreationEvent(identityProvider *miniov1alpha1.IdentityProvider) error {
	i.recorder.Event(identityProvider, event.Event{
		Type:    event.TypeNormal,
		Reason:  "Created",
		Message: "IdentityProvider successfully created",
	})
	return nil
}

func (i *identityProviderClient) createIdentityProvider(ctx context.Context, identityProvider *miniov1alpha1.IdentityProvider) error {
	return i.createOrUpdateIdentityProvider(ctx, identityProvider, false)
}

func (i *identityProviderClient) createOrUpdateIdentityProvider(ctx context.Context, identityProvider *miniov1alpha1.IdentityProvider, update bool) error {
	cfgType := madmin.OpenidIDPCfg

	name := identityProvider.GetIdentityProviderName()
	var input = []string{
		"client_id=" + identityProvider.Spec.ForProvider.ClientId,
		"client_secret=" + identityProvider.Spec.ForProvider.ClientSecret,
		"config_url=" + identityProvider.Spec.ForProvider.ConfigUrl,
		"scopes=" + identityProvider.Spec.ForProvider.Scopes,
		"redirect_uri=" + identityProvider.Spec.ForProvider.RedirectUrl,
		"display_name=" + identityProvider.Spec.ForProvider.DisplayName,
		"claim_name=" + identityProvider.Spec.ForProvider.ClaimName,
		"claim_userinfo=" + identityProvider.Spec.ForProvider.ClaimUserInfo,
		"redirect_uri_dynamic=" + identityProvider.Spec.ForProvider.RedirectUriDynamic,
	}

	cfgData := strings.Join(input, " ")

	restart, err := i.ma.AddOrUpdateIDPConfig(ctx, cfgType, name, cfgData, update)
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
