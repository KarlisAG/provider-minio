package group

import (
	"context"
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (g *groupClient) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("observing resource")

	groupResource, ok := mg.(*miniov1alpha1.Group)
	if !ok {
		return managed.ExternalObservation{}, errNotGroup
	}

	groups, err := g.ma.ListGroups(ctx)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	groupName := groupResource.GetGroupName()

	for _, group := range groups {
		if group == groupName {
			groupDescription, err := g.ma.GetGroupDescription(ctx, groupName)
			if err != nil {
				return managed.ExternalObservation{}, err
			}

			if groupDescription.Status == "disabled" {
				return managed.ExternalObservation{ResourceExists: true}, nil
			}

			if (len(groupResource.Spec.ForProvider.Users) > 0 || len(groupDescription.Members) > 0) && !g.usersMatch(groupResource, groupDescription.Members) {
				return managed.ExternalObservation{ResourceExists: true}, nil
			}

			if (len(groupResource.Spec.ForProvider.Policies) > 0 || groupDescription.Policy != "") && !g.policiesMatch(groupResource, groupDescription.Policy) {
				return managed.ExternalObservation{ResourceExists: true}, nil
			}

			groupResource.SetConditions(xpv1.Available())

			groupResource.Status.AtProvider.GroupName = groupName

			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
		}
	}

	return managed.ExternalObservation{}, nil
}

func (g *groupClient) usersMatch(group *miniov1alpha1.Group, users []string) bool {
	return g.sliceContentMatches(group.Spec.ForProvider.Users, users)
}

func (g *groupClient) policiesMatch(group *miniov1alpha1.Group, policy string) bool {
	// policy contains a string with all applied policies separated by comma
	policies := strings.Split(policy, ",")
	return g.sliceContentMatches(group.Spec.ForProvider.Policies, policies)
}

// While we could use reflect.DeepEqual, it relies on the provided list in CR and in MinIO to be in the same order
// It might be too much to ask from the user to provide the list in the same order as it would be in MinIO
func (g *groupClient) sliceContentMatches(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	sliceMap := make(map[string]bool)
	for _, item := range slice1 {
		sliceMap[item] = true
	}

	for _, item := range slice2 {
		if !sliceMap[item] {
			return false
		}
	}

	return true
}
