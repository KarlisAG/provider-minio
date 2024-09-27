package v1alpha1

import (
	"reflect"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	SchemeBuilder.Register(&IdentityProvider{}, &IdentityProviderList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="Synced",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,minio}
// +kubebuilder:webhook:verbs=create;update,path=/validate-minio-crossplane-io-v1alpha1-identityprovider,mutating=false,failurePolicy=fail,groups=minio.crossplane.io,resources=identityproviders,versions=v1alpha1,name=identityproviders.minio.crossplane.io,sideEffects=None,admissionReviewVersions=v1

type IdentityProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityProviderSpec   `json:"spec"`
	Status IdentityProviderStatus `json:"status,omitempty"`
}

type IdentityProviderSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ProviderReference *xpv1.Reference `json:"providerReference,omitempty"`

	ForProvider IdentityProviderParameters `json:"forProvider,omitempty"`
}

type IdentityProviderStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          IdentityProviderProviderStatus `json:"atProvider,omitempty"`
}

type IdentityProviderProviderStatus struct {
	// IdentityProvider is actual identity provider name created in MinIO.
	IdentityProvider string `json:"identityProvider,omitempty"`
}

type IdentityProviderParameters struct {
	// +kubebuilder:validation:Required
	// Specify the unique public identifier MinIO uses when authenticating user credentials against the OIDC compatible provider.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html#client-id
	ClientId string `json:"clientId,omitempty"`

	// +kubebuilder:validation:Required
	// Specify the client secret MinIO uses when authenticating user credentials against the OIDC compatible provider. This field may be optional depending on the provider.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html#client-secret
	ClientSecret string `json:"clientSecret,omitempty"`

	// +kubebuilder:validation:Required
	// Specify the URL for the OIDC compatible provider discovery document.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html#config-url
	ConfigUrl string `json:"configUrl,omitempty"`

	// +kubebuilder:default="openid,profile,email"
	// Scopes specify a comma-separated list of scopes.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html#scopes
	// Defaults to `openid,profile,email` if unset.
	Scopes string `json:"scopes,omitempty"`

	// Specify the Fully Qualified Domain Name (FQDN) the MinIO Console listens for incoming connections on.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/console.html#envvar.MINIO_BROWSER_REDIRECT_URL
	RedirectUrl string `json:"redirectUrl,omitempty"`

	// Name is the name of the identity provider to create.
	// Defaults to `metadata.name` if unset.
	Name string `json:"name,omitempty"`

	// Specify the user-facing name the MinIO Console displays on the login screen.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html#display-name
	DisplayName string `json:"displayName,omitempty"`

	// +kubebuilder:default="policy"
	// Specify the name of the JWT Claim MinIO uses to identify the policies to attach to the authenticated user.
	// The claim can contain one or more comma-separated policy names to attach to the user.
	// The claim must contain at least one policy for the user to have any permissions on the MinIO server.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html#claim-name
	// Defaults to `policy` if unset.
	ClaimName string `json:"claimName,omitempty"`

	// +kubebuilder:default="off"
	// Allow MinIO to fetch claims from the UserInfo Endpoint for the authenticated user.
	// Valid values are `on` or `off`.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html#user-info
	// Defaults to `off` if unset.
	ClaimUserInfo string `json:"claimUserInfo,omitempty"`

	// +kubebuilder:default="off"
	// The MinIO Console defaults to using the hostname of the node making the authentication request as part of the redirect URI provided to the OIDC provider. For MinIO deployments behind a load balancer using a round-robin protocol, this may result in the load balancer returning the response to a different MinIO Node than the originating client.
	// Specify this option as `on` to direct the MinIO Console to use the Host header of the originating request to construct the redirect URI passed to the OIDC provider.
	// See https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html#dynamic-uri-redirect
	// Defaults to `off`.
	RedirectUriDynamic string `json:"redirectUriDynamic,omitempty"`
}

// +kubebuilder:object:root=true

type IdentityProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityProvider `json:"items"`
}

// Dummy type metadata.
var (
	IdentityProviderKind             = reflect.TypeOf(IdentityProvider{}).Name()
	IdentityProviderGroupKind        = schema.GroupKind{Group: Group, Kind: IdentityProviderKind}.String()
	IdentityProviderKindAPIVersion   = IdentityProviderKind + "." + SchemeGroupVersion.String()
	IdentityProviderGroupVersionKind = SchemeGroupVersion.WithKind(IdentityProviderKind)
)

// GetIdentityProviderName returns the spec.forProvider.Name, if given, otherwise defaults to metadata.name.
func (in *IdentityProvider) GetIdentityProviderName() string {
	if in.Spec.ForProvider.Name == "" {
		return in.Name
	}
	return in.Spec.ForProvider.Name
}
