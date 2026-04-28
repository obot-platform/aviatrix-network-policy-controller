package controller

import (
	"context"
	"strings"
	"sync"
	"time"

	aviatrixv1alpha1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/networking.aviatrix.com/v1alpha1"
	obotv1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/obot.obot.ai/v1"
	"github.com/obot-platform/aviatrix-network-policy-controller/pkg/translate"
	"github.com/obot-platform/nah/pkg/apply"
	"github.com/obot-platform/nah/pkg/router"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type sourceTrigger interface {
	Trigger(ctx context.Context, gvk schema.GroupVersionKind, key string, delay time.Duration) error
}

type sourceKey struct {
	namespace string
	name      string
}

type FirewallPolicyWatcher struct {
	SourceTrigger    sourceTrigger
	SourceClient     kclient.Client
	RuntimeNamespace string
	SourceNamespace  string

	lock             sync.RWMutex
	sourceByFirewall map[string]sourceKey
}

func NewFirewallPolicyWatcher(sourceTrigger sourceTrigger, sourceClient kclient.Client, runtimeNamespace string, sourceNamespace ...string) *FirewallPolicyWatcher {
	w := &FirewallPolicyWatcher{
		SourceTrigger:    sourceTrigger,
		SourceClient:     sourceClient,
		RuntimeNamespace: runtimeNamespace,
		sourceByFirewall: map[string]sourceKey{},
	}
	if len(sourceNamespace) > 0 {
		w.SourceNamespace = sourceNamespace[0]
	}
	return w
}

func (w *FirewallPolicyWatcher) Handle(req router.Request, _ router.Response) error {
	if w.RuntimeNamespace != "" && req.Namespace != w.RuntimeNamespace {
		return nil
	}

	firewallKey := router.Key(req.Namespace, req.Name).String()
	source, ok, err := w.sourceForRequest(req.Ctx, firewallKey, req.Name, req.Object)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	return w.SourceTrigger.Trigger(req.Ctx, obotv1.SchemeGroupVersion.WithKind("MCPNetworkPolicy"), source.String(), 0)
}

func (w *FirewallPolicyWatcher) sourceForRequest(ctx context.Context, firewallKey, firewallName string, obj kclient.Object) (kclient.ObjectKey, bool, error) {
	if obj != nil {
		firewallPolicy, ok := obj.(*aviatrixv1alpha1.FirewallPolicy)
		if !ok {
			return kclient.ObjectKey{}, false, nil
		}
		if source, ok := sourceFromFirewallPolicy(firewallPolicy); ok {
			w.lock.Lock()
			w.sourceByFirewall[firewallKey] = source
			w.lock.Unlock()
			return router.Key(source.namespace, source.name), true, nil
		}
	}

	w.lock.RLock()
	source, ok := w.sourceByFirewall[firewallKey]
	w.lock.RUnlock()
	if ok {
		return router.Key(source.namespace, source.name), true, nil
	}

	if w.SourceNamespace != "" {
		if sourceName, ok := translate.MCPNetworkPolicyNameFromFirewallPolicyName(firewallName); ok {
			return router.Key(w.SourceNamespace, sourceName), true, nil
		}
	}

	return w.sourceByFirewallName(ctx, firewallName)
}

func (w *FirewallPolicyWatcher) sourceByFirewallName(ctx context.Context, firewallName string) (kclient.ObjectKey, bool, error) {
	if w.SourceClient == nil {
		return kclient.ObjectKey{}, false, nil
	}

	var policies obotv1.MCPNetworkPolicyList
	if err := w.SourceClient.List(ctx, &policies); err != nil {
		return kclient.ObjectKey{}, false, err
	}

	for _, policy := range policies.Items {
		if translate.NameForMCPNetworkPolicy(policy.Name) == firewallName {
			return router.Key(policy.Namespace, policy.Name), true, nil
		}
	}

	return kclient.ObjectKey{}, false, nil
}

func sourceFromFirewallPolicy(policy *aviatrixv1alpha1.FirewallPolicy) (sourceKey, bool) {
	if policy == nil {
		return sourceKey{}, false
	}

	subContext := policy.GetAnnotations()[apply.LabelSubContext]
	namespace, name, ok := parseSourceSubContext(subContext)
	if !ok {
		return sourceKey{}, false
	}

	return sourceKey{namespace: namespace, name: name}, true
}

func parseSourceSubContext(subContext string) (namespace, name string, ok bool) {
	value, ok := strings.CutPrefix(subContext, "mcp-network-policy/")
	if !ok {
		return "", "", false
	}

	namespace, name, ok = strings.Cut(value, "/")
	if !ok || namespace == "" || name == "" {
		return "", "", false
	}

	return namespace, name, true
}
