package controller

import (
	"fmt"

	aviatrixv1alpha1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/networking.aviatrix.com/v1alpha1"
	obotv1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/obot.obot.ai/v1"
	"github.com/obot-platform/aviatrix-network-policy-controller/pkg/translate"
	"github.com/obot-platform/nah/pkg/apply"
	"github.com/obot-platform/nah/pkg/router"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	RuntimeClient    kclient.Client
	RuntimeNamespace string
}

func (h *Handler) Reconcile(req router.Request, _ router.Response) error {
	app := apply.New(h.RuntimeClient).
		WithNamespace(h.RuntimeNamespace).
		WithOwnerSubContext(sourceSubContext(req.Namespace, req.Name)).
		WithPruneTypes(&aviatrixv1alpha1.FirewallPolicy{})

	if req.Object == nil {
		return app.Apply(req.Ctx, nil)
	}

	policy := req.Object.(*obotv1.MCPNetworkPolicy)
	desired := translate.ToFirewallPolicy(policy, h.RuntimeNamespace)
	if desired == nil {
		return app.Apply(req.Ctx, nil)
	}

	return app.Apply(req.Ctx, nil, desired)
}

func sourceSubContext(namespace, name string) string {
	return fmt.Sprintf("mcp-network-policy/%s/%s", namespace, name)
}
