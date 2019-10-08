package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	spannerhorizontalautoscalerv1alpha1 "github.com/micnncim/spanner-horizontal-autoscaler/api/v1alpha1"
	"github.com/micnncim/spanner-horizontal-autoscaler/pkg/pointer"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// Client is a client for Stackdriver Monitoring.
type Client struct {
	monitoringMetricClient *monitoring.MetricClient
	projectID              string
	syncPeriod             time.Duration
	log                    logr.Logger
}

type Option func(*Client)

func WithSyncPeriod(syncPeriod time.Duration) Option {
	return func(c *Client) {
		c.syncPeriod = syncPeriod
	}
}

func WithLog(log logr.Logger) Option {
	return func(c *Client) {
		c.log = log
	}
}

// NewClient returns a new Client.
func NewClient(ctx context.Context, projectID string, opts ...Option) (*Client, error) {
	metricClient, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, err
	}
	c := &Client{
		monitoringMetricClient: metricClient,
		projectID:              projectID,
		syncPeriod:             30 * time.Second,
		log:                    zapr.NewLogger(zap.NewNop()),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

type SpannerInstanceStatus struct {
	Name string
	spannerhorizontalautoscalerv1alpha1.SpannerInstanceStatus
}

// ReadMetrics reads time series metrics.
// https://cloud.google.com/monitoring/custom-metrics/reading-metrics?hl=ja#monitoring_read_timeseries_fields-go
func (c *Client) GetSpannerInstanceStatuses(ctx context.Context) ([]*SpannerInstanceStatus, error) {
	now := time.Now()
	startTime := now.UTC().Add(c.syncPeriod)
	endTime := now.UTC()
	filter := `
		metric.type = "spanner.googleapis.com/instance/cpu/utilization_by_priority" AND
		metric.label.priority = "high"
`

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", c.projectID),
		// TODO: Fix metrics type and enable to specify with argument.
		Filter: filter,
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: startTime.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endTime.Unix(),
			},
		},
		View: monitoringpb.ListTimeSeriesRequest_FULL,
	}

	var statuses []*SpannerInstanceStatus
	it := c.monitoringMetricClient.ListTimeSeries(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// v is CPU utilization.
		v := int32(resp.GetPoints()[0].GetValue().GetDoubleValue())
		statuses = append(statuses, &SpannerInstanceStatus{
			Name: resp.GetMetric().Labels["database"],
			SpannerInstanceStatus: spannerhorizontalautoscalerv1alpha1.SpannerInstanceStatus{
				CPUUtilization: pointer.Int32(v),
			},
		})
	}

	return statuses, nil
}
