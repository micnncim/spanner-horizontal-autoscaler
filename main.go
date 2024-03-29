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

package main

import (
	"context"
	"flag"
	"os"
	"time"

	spannerhorizontalautoscalerv1alpha1 "github.com/micnncim/spanner-horizontal-autoscaler/api/v1alpha1"
	"github.com/micnncim/spanner-horizontal-autoscaler/controllers"
	"github.com/micnncim/spanner-horizontal-autoscaler/pkg/monitoring"
	"github.com/micnncim/spanner-horizontal-autoscaler/pkg/pointer"
	"github.com/micnncim/spanner-horizontal-autoscaler/pkg/spanner"

	spanneradmin "cloud.google.com/go/spanner/admin/instance/apiv1"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = spannerhorizontalautoscalerv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		SyncPeriod:         pointer.Duration(30 * time.Second),
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	ctx := context.Background()

	spannerAdminClient, err := spanneradmin.NewInstanceAdminClient(ctx)
	if err != nil {
		setupLog.Error(err, "unable to create SpannerAdminClient")
		os.Exit(1)
	}
	spannerClient := spanner.NewClient(spannerAdminClient, spanner.WithLog(ctrl.Log.WithName("spanner")))

	// TODO: Set projectID appropriately.
	projectID := ""
	monitoringClient, err := monitoring.NewClient(ctx, projectID)
	if err != nil {
		setupLog.Error(err, "unable to create monitoring client")
		os.Exit(1)
	}

	if err = (&controllers.SpannerInstanceReconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("SpannerInstance"),
		Scheme:     mgr.GetScheme(),
		Recorder:   mgr.GetEventRecorderFor("spanner-horizontal-autoscaler"),
		Spanner:    spannerClient,
		Monitoring: monitoringClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SpannerInstance")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
