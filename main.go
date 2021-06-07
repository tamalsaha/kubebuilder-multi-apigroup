/*
Copyright 2021.

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
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	auditlib "go.bytebuilders.dev/audit/lib"
	_ "go.bytebuilders.dev/license-verifier/info"
	license "go.bytebuilders.dev/license-verifier/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"kmodules.xyz/client-go/discovery"
	"kmodules.xyz/client-go/tools/cli"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	chartv1alpha1 "github.com/tamalsaha/kubebuilder-multi-apigroup/apis/chart/v1alpha1"
	corev1alpha1 "github.com/tamalsaha/kubebuilder-multi-apigroup/apis/core/v1alpha1"
	modulev1alpha1 "github.com/tamalsaha/kubebuilder-multi-apigroup/apis/module/v1alpha1"
	chartcontrollers "github.com/tamalsaha/kubebuilder-multi-apigroup/controllers/chart"
	corecontrollers "github.com/tamalsaha/kubebuilder-multi-apigroup/controllers/core"
	modulecontrollers "github.com/tamalsaha/kubebuilder-multi-apigroup/controllers/module"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(modulev1alpha1.AddToScheme(scheme))
	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
	utilruntime.Must(chartv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var licenseFile string
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&licenseFile, "license-file", licenseFile, "Path to license file")
	flag.BoolVar(&cli.EnableAnalytics, "enable-analytics", cli.EnableAnalytics, "Send analytical events")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx := ctrl.SetupSignalHandler()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "5b87adeb.x-helm.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// audit event publisher
	var auditor *auditlib.EventPublisher
	if licenseFile != "" && cli.EnableAnalytics {
		cfg := mgr.GetConfig()
		kc, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			setupLog.Error(err, "unable to create Kubernetes client")
			os.Exit(1)
		}
		mapper := discovery.NewResourceMapper(mgr.GetRESTMapper())
		fn := auditlib.BillingEventCreator{
			Mapper: mapper,
		}
		auditor = auditlib.NewResilientEventPublisher(func() (*auditlib.NatsConfig, error) {
			return auditlib.NewNatsConfig(kc.CoreV1().Namespaces(), licenseFile)
		}, mapper, fn.CreateEvent)
	}

	if err = (&modulecontrollers.WorkflowReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("module").WithName("Workflow"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Workflow")
		os.Exit(1)
	}
	if err = auditor.SetupWithManager(ctx, mgr, &modulev1alpha1.Workflow{}); err != nil {
		setupLog.Error(err, "unable to set up auditor", "auditor", "Workflow")
		os.Exit(1)
	}

	if err = (&modulecontrollers.ActionReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("module").WithName("Action"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Action")
		os.Exit(1)
	}
	if err = (&corecontrollers.ReleaseReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("core").WithName("Release"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Release")
		os.Exit(1)
	}
	if err = (&chartcontrollers.RepositoryReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("chart").WithName("Repository"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Repository")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// Start periodic license verification
	//nolint:errcheck
	go license.VerifyLicensePeriodically(mgr.GetConfig(), licenseFile, ctx.Done())

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
