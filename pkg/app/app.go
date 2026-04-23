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
)

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	scheme := runtime.NewScheme()
	if err := obotv1.AddToScheme(scheme); err != nil {
		return err
	}
	if err := aviatrixv1alpha1.AddToScheme(scheme); err != nil {
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
		Namespace:   "default",
		HealthzPort: 8081,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage router: %w", err)
	}

	inCluster, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	runtimeClient, err := kclient.New(inCluster, kclient.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create runtime client: %w", err)
	}

	h := &controller.Handler{
		RuntimeClient:    runtimeClient,
		RuntimeNamespace: cfg.MCPRuntimeNamespace,
	}
	r.Type(&obotv1.MCPNetworkPolicy{}).IncludeRemoved().HandlerFunc(h.Reconcile)

	if err := r.Start(ctx); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return nil
	case <-r.Stopped():
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("storage router stopped")
	}
}
