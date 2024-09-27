package identityprovider

import (
	"context"
	"fmt"

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

	providerConfigRef := identityProvider.Spec.ProviderConfigReference
	if providerConfigRef == nil || providerConfigRef.Name == "" {
		return nil, fmt.Errorf(".spec.providerConfigRef.name is required")
	}
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator.
func (v *Validator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newIdentityProvider := newObj.(*miniov1alpha1.IdentityProvider)
	v.log.V(1).Info("Validate update")

	providerConfigRef := newIdentityProvider.Spec.ProviderConfigReference
	if providerConfigRef == nil || providerConfigRef.Name == "" {
		return nil, field.Invalid(field.NewPath("spec", "providerConfigRef", "name"), "null", "Provider config is required")
	}
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator.
func (v *Validator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	v.log.V(1).Info("validate delete (noop)")
	return nil, nil
}
