package group

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

func (g *groupClient) Delete(ctx context.Context, mg resource.Managed) error {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("deleting resource")

	group, ok := mg.(*miniov1alpha1.Group)
	if !ok {
		return errNotGroup
	}

	group.SetConditions(xpv1.Deleting())

	groupName := group.GetGroupName()
	groupDescription, err := g.ma.GetGroupDescription(ctx, groupName)
	if err != nil {
		return err
	}

	// We are passing all users that are currently in MinIO and not what we have declared
	// This is to avoid the potential situation, where just before deletion a user is manually added/deleted in the group and that mismatch could cause errors
	err = g.removeUsersFromGroup(ctx, groupName, groupDescription.Members)
	if err != nil {
		return err
	}

	// We need to run basically the same underlying function again, because a group can't be deleted if it has any users still attached to it
	// So we first delete all users then run the same function again to delete the group, as it shouldn't have any users attached to it anymore
	err = g.deleteGroup(ctx, groupName)
	if err != nil {
		return err
	}

	g.emitDeletionEvent(group)

	return nil
}

func (g *groupClient) emitDeletionEvent(group *miniov1alpha1.Group) error {
	g.recorder.Event(group, event.Event{
		Type:    event.TypeNormal,
		Reason:  "Deleted",
		Message: "Group successfully deleted",
	})
	return nil
}

// This function is used in update.go as it needs to remove users that are not declared in the CR
func (g *groupClient) removeUsersFromGroup(ctx context.Context, groupName string, users []string) error {
	return g.createUpdateOrDeleteGroup(ctx, groupName, users, true)
}

func (g *groupClient) deleteGroup(ctx context.Context, groupName string) error {
	return g.createUpdateOrDeleteGroup(ctx, groupName, nil, true)
}
