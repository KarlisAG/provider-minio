package group

import (
	"context"
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/minio/madmin-go/v3"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (g *groupClient) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("creating resource")

	group, ok := mg.(*miniov1alpha1.Group)
	if !ok {
		return managed.ExternalCreation{}, errNotGroup
	}

	group.SetConditions(xpv1.Creating())

	groupName := group.GetGroupName()

	err := g.createGroup(ctx, groupName, group.Spec.ForProvider.Users)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if len(group.Spec.ForProvider.Policies) > 0 {
		err = g.attachPolicy(ctx, group)
		if err != nil {
			return managed.ExternalCreation{}, err
		}
	}

	return managed.ExternalCreation{}, g.emitCreationEvent(group)
}

func (g *groupClient) emitCreationEvent(group *miniov1alpha1.Group) error {
	g.recorder.Event(group, event.Event{
		Type:    event.TypeNormal,
		Reason:  "Created",
		Message: "Group successfully created",
	})
	return nil
}

func (g *groupClient) createGroup(ctx context.Context, groupName string, users []string) error {
	return g.createUpdateOrDeleteGroup(ctx, groupName, users, false)
}

// To avoid repetitiveness, this function is called from create.go, update.go and delete.go
// That is because the imported madmin package only has a single underlying function that achieves such group (and user within) manipulation
func (g *groupClient) createUpdateOrDeleteGroup(ctx context.Context, groupName string, users []string, remove bool) error {
	groupAddRemove := madmin.GroupAddRemove{
		Group:    groupName,
		Members:  users,
		IsRemove: remove,
	}

	err := g.ma.UpdateGroupMembers(ctx, groupAddRemove)
	if err != nil {
		return err
	}
	return nil
}

// This function is also called from update.go because madmin doesn't have an option to update attached policies, only add or remove
func (g *groupClient) attachPolicy(ctx context.Context, group *miniov1alpha1.Group) error {
	groupName := group.GetGroupName()
	groupDescription, err := g.ma.GetGroupDescription(ctx, groupName)
	if err != nil {
		return err
	}

	var policiesToAttach []string
	// If the group has policies attached, we need to strip out the ones that are also declared in the CR to avoid errors
	if groupDescription.Policy != "" {
		policies := strings.Split(groupDescription.Policy, ",")
		policiesToAttach = stripExistingPolicies(policies, group.Spec.ForProvider.Policies)
	} else {
		policiesToAttach = group.Spec.ForProvider.Policies
	}

	policyRequest := madmin.PolicyAssociationReq{
		Group:    groupName,
		Policies: policiesToAttach,
	}

	_, err = g.ma.AttachPolicy(ctx, policyRequest)
	if err != nil {
		return err
	}
	return nil
}

// stripExistingPolicies returns a slice of policies that are not already attached to the group
// This is required because madmin.AttachPolicy() will return an error if a policy is already attached
func stripExistingPolicies(existingPolicies, newPolicies []string) []string {
	policyMap := make(map[string]bool)
	result := []string{}

	for _, policy := range existingPolicies {
		policyMap[policy] = true
	}

	// Check new policies against the map and add non-existing ones to the result
	for _, policy := range newPolicies {
		if !policyMap[policy] {
			result = append(result, policy)
		}
	}

	return result
}
