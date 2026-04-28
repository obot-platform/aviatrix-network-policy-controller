package controller

import (
	"fmt"

	aviatrixv1alpha1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/networking.aviatrix.com/v1alpha1"
	obotv1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/obot.obot.ai/v1"
	"github.com/obot-platform/aviatrix-network-policy-controller/pkg/translate"
	"github.com/obot-platform/nah/pkg/apply"
	"github.com/obot-platform/nah/pkg/router"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

var controllerLog = ctrlruntimelog.Log.WithName("controller")

type Handler struct {
	RuntimeClient    kclient.Client
	RuntimeNamespace string
}

func (h *Handler) Reconcile(req router.Request, _ router.Response) error {
	log := controllerLog.WithValues(
		"sourceNamespace", req.Namespace,
		"sourceName", req.Name,
		"runtimeNamespace", h.RuntimeNamespace)

	app := apply.New(h.RuntimeClient).
		WithNamespace(h.RuntimeNamespace).
		WithOwnerSubContext(sourceSubContext(req.Namespace, req.Name)).
		WithPruneTypes(&aviatrixv1alpha1.FirewallPolicy{})

	if req.Object == nil {
		if err := app.Apply(req.Ctx, nil); err != nil {
			log.Error(err, "failed to prune managed FirewallPolicy")
			return err
		}
		return nil
	}

	policy, ok := req.Object.(*obotv1.MCPNetworkPolicy)
	if !ok {
		err := fmt.Errorf("unexpected object type %T", req.Object)
		log.Error(err, "failed to reconcile MCPNetworkPolicy")
		return err
	}

	desired, err := translate.ToFirewallPolicy(policy, h.RuntimeNamespace)
	if err != nil {
		log.Error(err, "failed to translate MCPNetworkPolicy")
		return err
	}

	log = log.WithValues("firewallPolicyNamespace", desired.Namespace, "firewallPolicyName", desired.Name)
	if err := app.Apply(req.Ctx, nil, desired); err != nil {
		log.Error(err, "failed to apply managed FirewallPolicy")
		return err
	}
	return nil
}

func sourceSubContext(namespace, name string) string {
	return fmt.Sprintf("mcp-network-policy/%s/%s", namespace, name)
}
