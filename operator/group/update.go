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

	groupName := group.GetGroupName()
	groupDescription, err := g.ma.GetGroupDescription(ctx, groupName)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if groupDescription.Status == "disabled" {
		err = g.ma.SetGroupStatus(ctx, groupName, madmin.GroupEnabled)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	// To remove users that are not declared in the CR
	if excessUsers := g.nonMatchingSliceEntries(group.Spec.ForProvider.Users, groupDescription.Members); len(excessUsers) > 0 {
		err = g.removeUsersFromGroup(ctx, groupName, excessUsers)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	err = g.updateGroup(ctx, groupName, group.Spec.ForProvider.Users)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if !g.policiesMatch(group, groupDescription.Policy) {
		// To remove policies that are not declared in the CR
		err = g.detachIncorrectPolicies(ctx, group.Spec.ForProvider.Policies, groupDescription.Policy, groupName)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}

		if len(group.Spec.ForProvider.Policies) > 0 {
			err = g.attachPolicy(ctx, group)
			if err != nil {
				return managed.ExternalUpdate{}, err
			}
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

func (g *groupClient) updateGroup(ctx context.Context, groupName string, users []string) error {
	return g.createUpdateOrDeleteGroup(ctx, groupName, users, false)
}

func (g *groupClient) detachIncorrectPolicies(ctx context.Context, expectedPolicies []string, currentPolicies, groupName string) error {
	policies := strings.Split(currentPolicies, ",")

	policiesToDetach := g.nonMatchingSliceEntries(expectedPolicies, policies)
	if len(policiesToDetach) > 0 && policiesToDetach[0] != "" {
		policyRequest := madmin.PolicyAssociationReq{
			Group:    groupName,
			Policies: policiesToDetach,
		}

		_, err := g.ma.DetachPolicy(ctx, policyRequest)
		if err != nil {
			return err
		}
	}
	return nil
}

// nonMatchingSliceEntries returns a new slice of strings based on the currentSliceEntries that are not present in the expectedSliceEntries
// This is required for policy and user cleanup as functions provided by madmin allow adding policies or users
// But if anything from the list is already present, an error will be thrown
func (g *groupClient) nonMatchingSliceEntries(expectedSliceEntries, currentSliceEntries []string) []string {
	nonMatchingSliceEntries := []string{}
	for _, policy := range currentSliceEntries {
		if !slices.Contains(expectedSliceEntries, policy) {
			nonMatchingSliceEntries = append(nonMatchingSliceEntries, policy)
		}
	}
	return nonMatchingSliceEntries
}
