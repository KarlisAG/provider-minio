package identityprovider

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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

	_, ok = identityProvider.GetAnnotations()[IdentityProviderCreatedAnnotationKey]
	if !ok {
		// The identityProvider has not yet been created
		return managed.ExternalObservation{}, nil
	}

	cfgType := madmin.OpenidIDPCfg
	identityProviderList, err := i.ma.ListIDPConfig(ctx, cfgType)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot list identity providers")
	}

	identityProviderName := identityProvider.GetIdentityProviderName()
	for _, idp := range identityProviderList {
		if idp.Name == identityProviderName {
			upToDate, err := i.IDPConfigUpToDate(ctx, cfgType, identityProviderName, identityProvider)
			if err != nil {
				return managed.ExternalObservation{}, err
			}

			identityProvider.Status.AtProvider = miniov1alpha1.IdentityProviderProviderStatus{
				Name:               identityProviderName,
				ClaimName:          identityProvider.Spec.ForProvider.ClaimName,
				ClaimUserInfo:      identityProvider.Spec.ForProvider.ClaimUserInfo,
				ClientId:           identityProvider.Spec.ForProvider.ClientId,
				ConfigUrl:          identityProvider.Spec.ForProvider.ConfigUrl,
				DisplayName:        identityProvider.Spec.ForProvider.DisplayName,
				RedirectUrl:        identityProvider.Spec.ForProvider.RedirectUrl,
				RedirectUriDynamic: identityProvider.Spec.ForProvider.RedirectUriDynamic,
				Scopes:             identityProvider.Spec.ForProvider.Scopes,
			}

			identityProvider.SetConditions(xpv1.Available())

			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: upToDate}, nil
		}
	}

	return managed.ExternalObservation{}, nil
}

func (i *identityProviderClient) IDPConfigUpToDate(ctx context.Context, cfgType, cfgName string, identityProvider *miniov1alpha1.IdentityProvider) (bool, error) {
	// Returned config doesn't include client_secret
	idpConfig, err := i.ma.GetIDPConfig(ctx, cfgType, cfgName)
	if err != nil {
		return false, errors.Wrap(err, "cannot get identity provider config")
	}

	var requiredKeys = map[string]string{
		"claim_name":           identityProvider.Spec.ForProvider.ClaimName,
		"claim_userinfo":       identityProvider.Spec.ForProvider.ClaimUserInfo,
		"client_id":            identityProvider.Spec.ForProvider.ClientId,
		"config_url":           identityProvider.Spec.ForProvider.ConfigUrl,
		"display_name":         identityProvider.Spec.ForProvider.DisplayName,
		"enable":               "on", // Assuming "enable" should be "on" if not "off"
		"redirect_uri":         identityProvider.Spec.ForProvider.RedirectUrl,
		"redirect_uri_dynamic": identityProvider.Spec.ForProvider.RedirectUriDynamic,
		"scopes":               identityProvider.Spec.ForProvider.Scopes,
	}

	// The config we get from GetIDPConfig() includes only the fields that are set, not empty and not default
	// Because of that we need to add custom logic to mark such fields as not up to date, if we set it and it is not present in the retrieved config
	var presentKeys = map[string]bool{
		"claim_name":   false,
		"display_name": false,
		"redirect_uri": false,
	}

	for _, config := range idpConfig.Info {
		if expectedValue, exists := requiredKeys[config.Key]; exists {
			if config.Key == "enable" && config.Value == "off" {
				return false, nil
			}
			if config.Value != expectedValue {
				return false, nil
			}
			if _, track := presentKeys[config.Key]; track {
				presentKeys[config.Key] = true
			}
		}
	}

	if (identityProvider.Spec.ForProvider.ClaimName != "" && !presentKeys["claim_name"]) ||
		(identityProvider.Spec.ForProvider.DisplayName != "" && !presentKeys["display_name"]) ||
		(identityProvider.Spec.ForProvider.RedirectUrl != "" && !presentKeys["redirect_uri"]) {
		return false, nil
	}

	return true, nil
}
