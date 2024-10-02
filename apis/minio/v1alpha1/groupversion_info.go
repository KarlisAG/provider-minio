// +kubebuilder:object:generate=true
// +groupName=minio.crossplane.io
// +versionName=v1alpha1

// Package v1alpha1 contains the v1alpha1 group minio.crossplane.io resources of provider-minio.
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	// Using ApiGroup name instead of Group as it then collides with the type name
	ApiGroup = "minio.crossplane.io"
	Version  = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: ApiGroup, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)
