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
		if err := app.Apply(req.Ctx, nil); err != nil {
			return err
		}
		return nil
	}

	policy, ok := req.Object.(*obotv1.MCPNetworkPolicy)
	if !ok {
		return fmt.Errorf("unexpected object type %T", req.Object)
	}

	desired, err := translate.ToFirewallPolicy(policy, h.RuntimeNamespace)
	if err != nil {
		return err
	}

	if err := app.Apply(req.Ctx, nil, desired); err != nil {
		return err
	}
	return nil
}

func sourceSubContext(namespace, name string) string {
	return fmt.Sprintf("mcp-network-policy/%s/%s", namespace, name)
}
