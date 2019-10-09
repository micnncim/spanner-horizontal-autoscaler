package spanner

import (
	"context"
	"fmt"

	spanneradmin "cloud.google.com/go/spanner/admin/instance/apiv1"
	"github.com/go-logr/logr"
	"google.golang.org/api/iterator"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
)

type Client struct {
	spannerInstanceAdminClient *spanneradmin.InstanceAdminClient
	log                        logr.Logger
}

type Option func(*Client)

func WithLog(log logr.Logger) Option {
	return func(c *Client) {
		c.log = log
	}
}

func NewClient(spannerInstanceAdminClient *spanneradmin.InstanceAdminClient, opts ...Option) *Client {
	c := &Client{
		spannerInstanceAdminClient: spannerInstanceAdminClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) GetInstance(ctx context.Context, instanceID string) (*instancepb.Instance, error) {
	return c.spannerInstanceAdminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: instanceID,
	})
}

func (c *Client) ListInstances(ctx context.Context, projectID string) ([]*instancepb.Instance, error) {
	it := c.spannerInstanceAdminClient.ListInstances(ctx, &instancepb.ListInstancesRequest{
		Parent: fmt.Sprintf("projects/%s", projectID),
	})
	var instances []*instancepb.Instance
	for {
		instance, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (c *Client) IncreaseInstanceNodeCount(ctx context.Context, instanceID string, deltaNodeCount int) error {
	instance, err := c.spannerInstanceAdminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: instanceID,
	})
	if err != nil {
		return err
	}

	instance.NodeCount += int32(deltaNodeCount)
	_, err = c.spannerInstanceAdminClient.UpdateInstance(ctx, &instancepb.UpdateInstanceRequest{
		Instance: instance,
	})

	return err
}
