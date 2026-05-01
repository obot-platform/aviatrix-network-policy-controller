package app

import (
	"context"
	"fmt"

	aviatrixv1alpha1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/networking.aviatrix.com/v1alpha1"
	obotv1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/obot.obot.ai/v1"
	"github.com/obot-platform/aviatrix-network-policy-controller/pkg/config"
	"github.com/obot-platform/aviatrix-network-policy-controller/pkg/controller"
	"github.com/obot-platform/nah"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

const obotStorageNamespace = "default"

var appLog = ctrlruntimelog.Log.WithName("app")

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		appLog.Error(err, "failed to load configuration")
		return err
	}
	appLog.Info("loaded configuration",
		"obotStorageURL", cfg.ObotStorageURL,
		"obotStorageTokenFile", cfg.ObotStorageTokenFile,
		"sourceNamespace", obotStorageNamespace,
		"runtimeNamespace", cfg.MCPRuntimeNamespace)

	scheme := runtime.NewScheme()
	if err := obotv1.AddToScheme(scheme); err != nil {
		appLog.Error(err, "failed to register Obot API scheme")
		return err
	}
	if err := aviatrixv1alpha1.AddToScheme(scheme); err != nil {
		appLog.Error(err, "failed to register Aviatrix API scheme")
		return err
	}

	storageConfig := &rest.Config{
		Host:            cfg.ObotStorageURL,
		BearerTokenFile: cfg.ObotStorageTokenFile,
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json",
			ContentType:        "application/json",
		},
		TLSClientConfig: rest.TLSClientConfig{Insecure: true}, // The controller talks to Obot over HTTPS, but skips TLS verification for in-cluster access.
	}

	r, err := nah.NewRouter("aviatrix-network-policy-controller", &nah.Options{
		RESTConfig:  storageConfig,
		Scheme:      scheme,
		Namespace:   obotStorageNamespace,
		HealthzPort: 8081,
	})
	if err != nil {
		appLog.Error(err, "failed to create storage router")
		return fmt.Errorf("failed to create storage router: %w", err)
	}
	appLog.Info("created storage router", "namespace", obotStorageNamespace, "healthzPort", 8081)

	inCluster, err := rest.InClusterConfig()
	if err != nil {
		appLog.Error(err, "failed to load in-cluster config")
		return fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	runtimeClient, err := kclient.New(inCluster, kclient.Options{Scheme: scheme})
	if err != nil {
		appLog.Error(err, "failed to create runtime client")
		return fmt.Errorf("failed to create runtime client: %w", err)
	}
	appLog.Info("created runtime client", "runtimeNamespace", cfg.MCPRuntimeNamespace)

	h := &controller.Handler{
		RuntimeClient:    runtimeClient,
		RuntimeNamespace: cfg.MCPRuntimeNamespace,
	}
	r.Type(&obotv1.MCPNetworkPolicy{}).IncludeRemoved().HandlerFunc(h.Reconcile)

	runtimeRouter, err := nah.NewRouter("aviatrix-firewallpolicy-watcher", &nah.Options{
		RESTConfig:  inCluster,
		Scheme:      scheme,
		Namespace:   cfg.MCPRuntimeNamespace,
		HealthzPort: -1,
	})
	if err != nil {
		appLog.Error(err, "failed to create runtime router")
		return fmt.Errorf("failed to create runtime router: %w", err)
	}
	firewallPolicyWatcher := controller.NewFirewallPolicyWatcher(r.Backend(), r.Backend(), cfg.MCPRuntimeNamespace, obotStorageNamespace)
	runtimeRouter.Type(&aviatrixv1alpha1.FirewallPolicy{}).IncludeRemoved().HandlerFunc(firewallPolicyWatcher.Handle)

	appLog.Info("starting storage router", "namespace", obotStorageNamespace)
	if err := r.Start(ctx); err != nil {
		appLog.Error(err, "failed to start storage router")
		return err
	}
	appLog.Info("starting runtime router", "namespace", cfg.MCPRuntimeNamespace)
	if err := runtimeRouter.Start(ctx); err != nil {
		appLog.Error(err, "failed to start runtime router")
		return err
	}
	appLog.Info("aviatrix network policy controller started successfully",
		"sourceNamespace", obotStorageNamespace,
		"runtimeNamespace", cfg.MCPRuntimeNamespace,
		"healthzPort", 8081)

	select {
	case <-ctx.Done():
		appLog.Info("controller context canceled")
		return nil
	case <-r.Stopped():
		if ctx.Err() != nil {
			appLog.Info("storage router stopped after context cancellation")
			return nil
		}
		appLog.Error(fmt.Errorf("storage router stopped"), "storage router stopped unexpectedly")
		return fmt.Errorf("storage router stopped")
	case <-runtimeRouter.Stopped():
		if ctx.Err() != nil {
			appLog.Info("runtime router stopped after context cancellation")
			return nil
		}
		appLog.Error(fmt.Errorf("runtime router stopped"), "runtime router stopped unexpectedly")
		return fmt.Errorf("runtime router stopped")
	}
}
