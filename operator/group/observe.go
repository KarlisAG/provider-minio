package group

import (
	"context"
	"reflect"
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

			if !g.usersMatch(groupResource, groupDescription.Members) {
				return managed.ExternalObservation{ResourceExists: true}, nil
			}

			if !g.policiesMatch(groupResource, groupDescription.Policy) {
				return managed.ExternalObservation{ResourceExists: true}, nil
			}

			groupResource.SetConditions(xpv1.Available())

			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
		}
	}

	return managed.ExternalObservation{}, nil
}

func (g *groupClient) usersMatch(group *miniov1alpha1.Group, users []string) bool {
	return reflect.DeepEqual(users, group.Spec.ForProvider.Users)
}

func (g *groupClient) policiesMatch(group *miniov1alpha1.Group, policy string) bool {
	// policy contains a string with all applied policies separated by comma
	policies := strings.Split(policy, ",")
	return reflect.DeepEqual(policies, group.Spec.ForProvider.Policies)
}
