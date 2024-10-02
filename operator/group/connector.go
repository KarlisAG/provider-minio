package group

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/minio/madmin-go/v3"
	miniov1alpha1 "github.com/vshn/provider-minio/apis/minio/v1alpha1"
	providerv1 "github.com/vshn/provider-minio/apis/provider/v1"
	"github.com/vshn/provider-minio/operator/minioutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errNotGroup = fmt.Errorf("managed resource is not a group")
)

type connector struct {
	kube     client.Client
	recorder event.Recorder
	usage    resource.Tracker
}

type groupClient struct {
	ma       *madmin.AdminClient
	recorder event.Recorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(1).Info("connecting resource")

	err := c.usage.Track(ctx, mg)
	if err != nil {
		return nil, err
	}

	group, ok := mg.(*miniov1alpha1.Group)
	if !ok {
		return nil, errNotGroup
	}

	config, err := c.getProviderConfig(ctx, group)
	if err != nil {
		return nil, err
	}

	ma, err := minioutil.NewMinioAdmin(ctx, c.kube, config)
	if err != nil {
		return nil, err
	}

	gc := &groupClient{
		ma:       ma,
		recorder: c.recorder,
	}

	return gc, nil
}

func (c *connector) getProviderConfig(ctx context.Context, group *miniov1alpha1.Group) (*providerv1.ProviderConfig, error) {
	configName := group.GetProviderConfigReference().Name
	config := &providerv1.ProviderConfig{}
	err := c.kube.Get(ctx, client.ObjectKey{Name: configName}, config)
	return config, err
}
