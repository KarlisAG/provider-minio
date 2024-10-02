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

	err := g.deleteGroup(ctx, group)
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

func (g *groupClient) deleteGroup(ctx context.Context, group *miniov1alpha1.Group) error {
	return g.createUpdateOrDeleteGroup(ctx, group, true)
}
