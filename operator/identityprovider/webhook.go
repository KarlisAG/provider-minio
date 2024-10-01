package identityprovider

import (
	"context"
	"fmt"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/go-logr/logr"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.CustomValidator = &Validator{}

// Validator validates admission requests.
type Validator struct {
	log logr.Logger
}

// ValidateCreate implements admission.CustomValidator.
func (v *Validator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	identityProvider, ok := obj.(*miniov1alpha1.IdentityProvider)
	if !ok {
		return nil, errNotIdentityProvider
	}
	name := identityProvider.GetIdentityProviderName()
	v.log.V(1).Info("Validate create", "name", name)

	return nil, v.validateIdentityProvider(identityProvider)
}

// ValidateUpdate implements admission.CustomValidator.
func (v *Validator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newIdentityProvider, ok := newObj.(*miniov1alpha1.IdentityProvider)
	if !ok {
		return nil, errNotIdentityProvider
	}
	v.log.V(1).Info("Validate update")

	return nil, v.validateIdentityProvider(newIdentityProvider)
}

// ValidateDelete implements admission.CustomValidator.
func (v *Validator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	v.log.V(1).Info("validate delete (noop)")
	return nil, nil
}

func (v *Validator) validateIdentityProvider(identityProvider *miniov1alpha1.IdentityProvider) error {
	if identityProvider.Spec.ForProvider.ClientSecret != "" && identityProvider.Spec.ForProvider.ClientSecretRef != (xpv1.SecretKeySelector{}) {
		return fmt.Errorf(".spec.forProvider.clientSecret and .spec.forProvider.clientSecretRef are mutual exclusive, please only specify one")
	}

	providerConfigRef := identityProvider.Spec.ProviderConfigReference
	if providerConfigRef == nil || providerConfigRef.Name == "" {
		return field.Invalid(field.NewPath("spec", "providerConfigRef", "name"), "null", "Provider config is required")
	}
	return nil
}
