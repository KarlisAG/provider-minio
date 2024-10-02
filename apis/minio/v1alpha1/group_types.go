package v1alpha1

import (
	"reflect"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="Synced",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="External Name",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,minio}
// +kubebuilder:webhook:verbs=create;update,path=/validate-minio-crossplane-io-v1alpha1-group,mutating=false,failurePolicy=fail,groups=minio.crossplane.io,resources=groups,versions=v1alpha1,name=groups.minio.crossplane.io,sideEffects=None,admissionReviewVersions=v1

type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec"`
	Status GroupStatus `json:"status,omitempty"`
}

type GroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ProviderReference *xpv1.Reference `json:"providerReference,omitempty"`

	ForProvider GroupParameters `json:"forProvider,omitempty"`
}

type GroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupProviderStatus `json:"atProvider,omitempty"`
}

type GroupProviderStatus struct {
	// Name of the group
	Group string `json:"policy,omitempty"`
}

type GroupParameters struct {
	// Name of the group to create.
	// Defaults to `metadata.name` if unset.
	Name string `json:"name,omitempty"`

	// List of users to add to the group.
	// They must exist before the group creation.
	// See: https://min.io/docs/minio/linux/reference/minio-mc-admin/mc-admin-group.html#mc.admin.group.add.MEMBERS
	Users []string `json:"users,omitempty"`

	// List of policy names to attach to the group.
	// They must exist before the group creation.
	// See: https://min.io/docs/minio/linux/reference/minio-mc-admin/mc-admin-policy-attach.html#command-mc.admin.policy.attach
	Policies []string `json:"policies,omitempty"`
}

// +kubebuilder:object:root=true

type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

// Dummy type metadata.
var (
	GroupKind             = reflect.TypeOf(Group{}).Name()
	GroupGroupKind        = schema.GroupKind{Group: ApiGroup, Kind: GroupKind}.String()
	GroupKindAPIVersion   = GroupKind + "." + SchemeGroupVersion.String()
	GroupGroupVersionKind = SchemeGroupVersion.WithKind(GroupKind)
)

func (in *Group) GetGroupName() string {
	if in.Spec.ForProvider.Name == "" {
		return in.Name
	}
	return in.Spec.ForProvider.Name
}
