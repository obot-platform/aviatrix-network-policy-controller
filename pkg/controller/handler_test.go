package controller

import (
	"testing"

	aviatrixv1alpha1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/networking.aviatrix.com/v1alpha1"
	obotv1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/obot.obot.ai/v1"
	"github.com/obot-platform/nah/pkg/apply"
	nahrouter "github.com/obot-platform/nah/pkg/router"
	"github.com/obot-platform/nah/pkg/router/tester"
	metarest "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := obotv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := aviatrixv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return scheme
}

func newHandler(t *testing.T, objs ...kclient.Object) *Handler {
	t.Helper()
	scheme := newScheme(t)
	mapper := metarest.NewDefaultRESTMapper([]schema.GroupVersion{
		obotv1.SchemeGroupVersion,
		aviatrixv1alpha1.SchemeGroupVersion,
	})
	mapper.Add(obotv1.SchemeGroupVersion.WithKind("MCPNetworkPolicy"), metarest.RESTScopeNamespace)
	mapper.Add(aviatrixv1alpha1.SchemeGroupVersion.WithKind("FirewallPolicy"), metarest.RESTScopeNamespace)

	return &Handler{
		RuntimeClient:    fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(objs...).Build(),
		RuntimeNamespace: "obot-mcp",
	}
}

func TestHandlerCreatesManagedFirewallPolicy(t *testing.T) {
	scheme := newScheme(t)
	policy := &obotv1.MCPNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-a", Namespace: "default", UID: "uid-a"},
		Spec: obotv1.MCPNetworkPolicySpec{
			MCPServerName: "server-a",
			PodSelector:   map[string]string{"app": "server-a"},
			EgressDomains: []string{"api.example.com"},
		},
	}

	handler := newHandler(t)
	if _, err := (&tester.Harness{Scheme: scheme}).InvokeFunc(t, policy, handler.Reconcile); err != nil {
		t.Fatal(err)
	}

	var created aviatrixv1alpha1.FirewallPolicy
	if err := handler.RuntimeClient.Get(t.Context(), kclient.ObjectKey{Namespace: "obot-mcp", Name: "obot-server-a-fw"}, &created); err != nil {
		t.Fatal(err)
	}
	if created.Labels["obot.ai/mcp-server-name"] != "server-a" {
		t.Fatalf("unexpected labels: %v", created.Labels)
	}
}

func TestHandlerLifecycle(t *testing.T) {
	scheme := newScheme(t)
	policy := &obotv1.MCPNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-a", Namespace: "default", UID: "uid-a"},
		Spec: obotv1.MCPNetworkPolicySpec{
			MCPServerName: "server-a",
			PodSelector:   map[string]string{"app": "server-a"},
			EgressDomains: []string{"api.example.com"},
		},
	}

	handler := newHandler(t)
	if _, err := (&tester.Harness{Scheme: scheme}).InvokeFunc(t, policy, handler.Reconcile); err != nil {
		t.Fatal(err)
	}

	var created aviatrixv1alpha1.FirewallPolicy
	if err := handler.RuntimeClient.Get(t.Context(), kclient.ObjectKey{Namespace: "obot-mcp", Name: "obot-server-a-fw"}, &created); err != nil {
		t.Fatal(err)
	}

	policy.Spec.EgressDomains = []string{"new.example.com"}
	if _, err := (&tester.Harness{Scheme: scheme}).InvokeFunc(t, policy, handler.Reconcile); err != nil {
		t.Fatal(err)
	}
	if err := handler.RuntimeClient.Get(t.Context(), kclient.ObjectKey{Namespace: "obot-mcp", Name: "obot-server-a-fw"}, &created); err != nil {
		t.Fatal(err)
	}
	if created.Spec.WebGroups[0].Domains[0] != "new.example.com" {
		t.Fatalf("expected updated domain, got %v", created.Spec.WebGroups[0].Domains)
	}

	policy.Spec.MCPServerName = "server-b"
	if _, err := (&tester.Harness{Scheme: scheme}).InvokeFunc(t, policy, handler.Reconcile); err != nil {
		t.Fatal(err)
	}
	if err := handler.RuntimeClient.Get(t.Context(), kclient.ObjectKey{Namespace: "obot-mcp", Name: "obot-server-b-fw"}, &created); err != nil {
		t.Fatal(err)
	}
	if err := handler.RuntimeClient.Get(t.Context(), kclient.ObjectKey{Namespace: "obot-mcp", Name: "obot-server-a-fw"}, &aviatrixv1alpha1.FirewallPolicy{}); err == nil {
		t.Fatal("expected stale FirewallPolicy to be pruned after rename")
	}

	policy.Spec.EgressDomains = nil
	policy.Spec.DenyAllEgress = false
	if _, err := (&tester.Harness{Scheme: scheme}).InvokeFunc(t, policy, handler.Reconcile); err != nil {
		t.Fatal(err)
	}
	if err := handler.RuntimeClient.Get(t.Context(), kclient.ObjectKey{Namespace: "obot-mcp", Name: "obot-server-a-fw"}, &created); err == nil {
		t.Fatal("expected FirewallPolicy to be deleted when policy is unenforced")
	}
}

func TestHandlerDeletesManagedFirewallPolicyWhenSourceRemoved(t *testing.T) {
	scheme := newScheme(t)
	labels, annotations, err := apply.GetLabelsAndAnnotations(scheme, sourceSubContext("default", "policy-a"), nil)
	if err != nil {
		t.Fatal(err)
	}

	existing := &aviatrixv1alpha1.FirewallPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "obot-server-a-fw",
			Namespace:   "obot-mcp",
			Annotations: annotations,
			Labels:      labels,
		},
	}
	handler := newHandler(t, existing)
	req := nahrouter.Request{
		Ctx:       t.Context(),
		Name:      "policy-a",
		Namespace: "default",
	}
	if err := handler.Reconcile(req, &nahrouter.ResponseWrapper{}); err != nil {
		t.Fatal(err)
	}

	if err := handler.RuntimeClient.Get(t.Context(), kclient.ObjectKey{Namespace: "obot-mcp", Name: "obot-server-a-fw"}, &aviatrixv1alpha1.FirewallPolicy{}); err == nil {
		t.Fatal("expected FirewallPolicy to be deleted on source removal")
	}
}
