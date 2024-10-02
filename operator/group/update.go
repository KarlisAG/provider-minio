package group

import (
	"context"
	"slices"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/minio/madmin-go/v3"
	miniov1 "github.com/vshn/provider-minio/apis/minio/v1"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (g *groupClient) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("updating resource")

	group, ok := mg.(*miniov1alpha1.Group)
	if !ok {
		return managed.ExternalUpdate{}, errNotGroup
	}

	group.SetConditions(miniov1.Updating())

	err := g.updateGroup(ctx, group)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	groupName := group.GetGroupName()
	groupDescription, err := g.ma.GetGroupDescription(ctx, groupName)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if !g.policiesMatch(group, groupDescription.Policy) {
		err = g.detachIncorrectPolicies(ctx, group.Spec.ForProvider.Policies, groupDescription.Policy, groupName)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}

		err = g.attachPolicy(ctx, group)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, g.emitUpdateEvent(group)
}

func (g *groupClient) emitUpdateEvent(group *miniov1alpha1.Group) error {
	g.recorder.Event(group, event.Event{
		Type:    event.TypeNormal,
		Reason:  "Updated",
		Message: "Group successfully updated",
	})
	return nil
}

func (g *groupClient) updateGroup(ctx context.Context, group *miniov1alpha1.Group) error {
	return g.createUpdateOrDeleteGroup(ctx, group, false)
}

func (g *groupClient) detachIncorrectPolicies(ctx context.Context, expectedPolicies []string, currentPolicies, groupName string) error {
	policies := strings.Split(currentPolicies, ",")

	policiesToDetach := g.nonMatchingPolicies(expectedPolicies, policies)

	policyRequest := madmin.PolicyAssociationReq{
		Group:    groupName,
		Policies: policiesToDetach,
	}
	g.ma.DetachPolicy(ctx, policyRequest)
	return nil
}

func (g *groupClient) nonMatchingPolicies(expectedPolicies, currentPolicies []string) []string {
	nonMatchingPolicies := []string{}
	for _, policy := range currentPolicies {
		if !slices.Contains(expectedPolicies, policy) {
			nonMatchingPolicies = append(nonMatchingPolicies, policy)
		}
	}
	return nonMatchingPolicies
}
