/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	spannerhorizontalautoscalerv1alpha1 "github.com/micnncim/spanner-horizontal-autoscaler/api/v1alpha1"
	"github.com/micnncim/spanner-horizontal-autoscaler/pkg/monitoring"
	"github.com/micnncim/spanner-horizontal-autoscaler/pkg/pointer"
	"github.com/micnncim/spanner-horizontal-autoscaler/pkg/spanner"
)

// SpannerInstanceReconciler reconciles a SpannerInstance object
type SpannerInstanceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	Spanner    *spanner.Client
	Monitoring *monitoring.Client
}

// +kubebuilder:rbac:groups=spannerhorizontalautoscaler.k8s.io,resources=spannerinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=spannerhorizontalautoscaler.k8s.io,resources=spannerinstances/status,verbs=get;update;patch

func (r *SpannerInstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("spannerinstance", req.NamespacedName)

	var spannerInstance spannerhorizontalautoscalerv1alpha1.SpannerInstance
	if err := r.Get(ctx, req.NamespacedName, &spannerInstance); err != nil {
		log.Error(err, "unable to fetch SpannerInstance")
		return ctrl.Result{}, err
	}

	if *spannerInstance.Status.CPUUtilization > *spannerInstance.Spec.CPUUtilizationThreshold &&
		*spannerInstance.Status.AvailableNodes < *spannerInstance.Spec.MaxNodes {
		// Scale node count.
		// TODO: Adopt suitable algorithm.
		// ref) https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/horizontal-pod-autoscaler.md
		// TODO: Reduce node count if necessary.
		if err := r.Spanner.UpdateInstanceNodeCount(ctx, spannerInstance.Spec.InstanceId, 1); err != nil {
			log.Error(err, "unable to update spanner instance node count")
			return ctrl.Result{}, err
		}
	}

	instanceID := spannerInstance.Spec.InstanceId

	// TODO: Set projectID appropriately.
	instance, err := r.Spanner.GetInstance(ctx, instanceID)
	if err != nil {
		log.Error(err, "unable to get spanner instance")
		return ctrl.Result{}, nil
	}
	spannerInstance.Status.AvailableNodes = pointer.Int32(instance.NodeCount)

	monitoringStatus, err := r.Monitoring.GetSpannerInstanceStatus(ctx, instanceID)
	if err != nil {
		log.Error(err, "unable to get spanner instance monitoringStatus")
		return ctrl.Result{}, nil
	}
	spannerInstance.Status.CPUUtilization = monitoringStatus.CPUUtilization

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &spannerInstance, func() error {
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create or update SpannerInstance")
		return ctrl.Result{}, nil
	}

	r.Recorder.Eventf(
		&spannerInstance,
		corev1.EventTypeNormal,
		"Updated",
		"Updated SpannerInstance",
	)

	return ctrl.Result{}, nil
}

func (r *SpannerInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&spannerhorizontalautoscalerv1alpha1.SpannerInstance{}).
		Complete(r)
}
