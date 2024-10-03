package group

import (
	"context"

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
	group, ok := obj.(*miniov1alpha1.Group)
	if !ok {
		return nil, errNotGroup
	}
	name := group.GetGroupName()
	v.log.V(1).Info("Validate create", "name", name)

	return nil, v.validateGroup(group)
}

// ValidateUpdate implements admission.CustomValidator.
func (v *Validator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newGroup, ok := newObj.(*miniov1alpha1.Group)
	if !ok {
		return nil, errNotGroup
	}
	oldGroup, ok := oldObj.(*miniov1alpha1.Group)
	if !ok {
		return nil, errNotGroup
	}

	// It is not forbidden to change the group name
	// It is blocked here because, if it is changed then in MinIO a new group will appear, but the old one won't get deleted
	if newGroup.GetGroupName() != oldGroup.GetGroupName() {
		return nil, field.Invalid(field.NewPath("spec", "forProvider", "name"), newGroup.GetGroupName(), "Changing the group name is not allowed")
	}
	v.log.V(1).Info("Validate update")

	return nil, v.validateGroup(newGroup)
}

// ValidateDelete implements admission.CustomValidator.
func (v *Validator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	v.log.V(1).Info("validate delete (noop)")
	return nil, nil
}

func (v *Validator) validateGroup(group *miniov1alpha1.Group) error {
	providerConfigRef := group.Spec.ProviderConfigReference
	if providerConfigRef == nil || providerConfigRef.Name == "" {
		return field.Invalid(field.NewPath("spec", "providerConfigRef", "name"), "null", "Provider config is required")
	}
	return nil
}
