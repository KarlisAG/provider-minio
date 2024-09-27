package identityprovider

import (
	"context"
	"strings"

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
	cfgType := madmin.OpenidIDPCfg

	name := identityProvider.GetIdentityProviderName()
	var input []string

	if identityProvider.Spec.ForProvider.ClientId != "" {
		clientId := "client_id=" + identityProvider.Spec.ForProvider.ClientId
		input = append(input, clientId)
	}
	if identityProvider.Spec.ForProvider.ClientSecret != "" {
		clientSecret := "client_secret=" + identityProvider.Spec.ForProvider.ClientSecret
		input = append(input, clientSecret)
	}
	if identityProvider.Spec.ForProvider.ConfigUrl != "" {
		configUrl := "config_url=" + identityProvider.Spec.ForProvider.ConfigUrl
		input = append(input, configUrl)
	}
	if identityProvider.Spec.ForProvider.Scopes != "" {
		scopes := "scopes=" + identityProvider.Spec.ForProvider.Scopes
		input = append(input, scopes)
	}
	if identityProvider.Spec.ForProvider.RedirectUrl != "" {
		redirectUrl := "redirect_uri=" + identityProvider.Spec.ForProvider.RedirectUrl
		input = append(input, redirectUrl)
	}
	if identityProvider.Spec.ForProvider.DisplayName != "" {
		displayName := "display_name=" + identityProvider.Spec.ForProvider.DisplayName
		input = append(input, displayName)
	}
	if identityProvider.Spec.ForProvider.ClaimName != "" {
		claimName := "claim_name=" + identityProvider.Spec.ForProvider.ClaimName
		input = append(input, claimName)
	}
	if identityProvider.Spec.ForProvider.ClaimUserInfo != "" {
		claimUserInfo := "claim_userinfo=" + identityProvider.Spec.ForProvider.ClaimUserInfo
		input = append(input, claimUserInfo)
	}
	if identityProvider.Spec.ForProvider.RedirectUriDynamic != "" {
		redirectUriDynamic := "redirect_uri_dynamic=" + identityProvider.Spec.ForProvider.RedirectUriDynamic
		input = append(input, redirectUriDynamic)
	}

	cfgData := strings.Join(input, " ")

	restart, err := i.ma.AddOrUpdateIDPConfig(ctx, cfgType, name, cfgData, false)
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
