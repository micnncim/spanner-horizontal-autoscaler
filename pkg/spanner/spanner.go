package spanner

import (
	"context"

	spanneradmin "cloud.google.com/go/spanner/admin/instance/apiv1"
	"github.com/go-logr/logr"
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

func (c *Client) UpdateInstanceNodeCount(ctx context.Context, instanceID string, nodeCount int) error {
	instance, err := c.spannerInstanceAdminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: instanceID,
	})
	if err != nil {
		return err
	}

	instance.NodeCount = int32(nodeCount)
	_, err = c.spannerInstanceAdminClient.UpdateInstance(ctx, &instancepb.UpdateInstanceRequest{
		Instance: instance,
	})

	return err
}
