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

	err := g.createGroup(ctx, group)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	err = g.attachPolicy(ctx, group)
	if err != nil {
		return managed.ExternalCreation{}, err
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

func (g *groupClient) createGroup(ctx context.Context, group *miniov1alpha1.Group) error {
	return g.createUpdateOrDeleteGroup(ctx, group, false)
}

func (g *groupClient) createUpdateOrDeleteGroup(ctx context.Context, group *miniov1alpha1.Group, remove bool) error {
	groupName := group.GetGroupName()
	groupAddRemove := madmin.GroupAddRemove{
		Group:    groupName,
		Members:  group.Spec.ForProvider.Users,
		IsRemove: remove,
	}

	err := g.ma.UpdateGroupMembers(ctx, groupAddRemove)
	if err != nil {
		return err
	}
	return nil
}

func (g *groupClient) attachPolicy(ctx context.Context, group *miniov1alpha1.Group) error {
	groupName := group.GetGroupName()
	groupDescription, err := g.ma.GetGroupDescription(ctx, groupName)
	if err != nil {
		return err
	}
	log := controllerruntime.LoggerFrom(ctx)
	var policiesToAttach []string
	log.Info("groupDescription.Policy: ", "groupDescription.Policy", groupDescription.Policy)
	log.Info("group.Spec.ForProvider.Policies: ", "group.Spec.ForProvider.Policies", group.Spec.ForProvider.Policies)
	if groupDescription.Policy != "" {
		policies := strings.Split(groupDescription.Policy, ",")
		log.Info("policies is not empty")
		// Stripping out policies that already exist to avoid errors
		policiesToAttach = stripExistingPolicies(policies, group.Spec.ForProvider.Policies)
	} else {
		log.Info("policies is empty")
		policiesToAttach = group.Spec.ForProvider.Policies
	}
	policyRequest := madmin.PolicyAssociationReq{
		Group:    groupName,
		Policies: policiesToAttach,
	}

	log.Info("Attaching policies to group")
	log.Info("Policies to attach: ", "policies", policiesToAttach)

	_, err = g.ma.AttachPolicy(ctx, policyRequest)
	if err != nil {
		return err
	}
	return nil
}

func stripExistingPolicies(existingPolicies, newPolicies []string) []string {
	policyMap := make(map[string]bool)
	result := []string{}

	// Add existing policies to the map
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
