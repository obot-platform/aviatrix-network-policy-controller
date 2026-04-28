package translate

import (
	"strings"
	"testing"

	aviatrixv1alpha1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/networking.aviatrix.com/v1alpha1"
	obotv1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/obot.obot.ai/v1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTranslateApprovedDomains(t *testing.T) {
	policy := &obotv1.MCPNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-a", UID: "uid-a"},
		Spec: obotv1.MCPNetworkPolicySpec{
			MCPServerName: "server-a",
			PodSelector:   map[string]string{"app": "server-a"},
			EgressDomains: []string{"b.example.com", "*.google.com", "a.example.com"},
		},
	}

	fp := ToFirewallPolicy(policy, "obot-mcp")
	require.NotNil(t, fp)
	require.Equal(t, "obot-policy-a-fw", fp.Name)
	require.Len(t, fp.Spec.Rules, 2)
	require.Equal(t, "allow-approved-egress", fp.Spec.Rules[0].Name)
	require.Equal(t, "deny-all-external", fp.Spec.Rules[1].Name)
	require.Equal(t, []string{"*.google.com", "a.example.com", "b.example.com"}, fp.Spec.WebGroups[0].Domains)
	require.Equal(t, map[string]string{"app": "server-a"}, fp.Spec.Rules[0].Selector.MatchLabels)
}

func TestTranslateDenyOnly(t *testing.T) {
	policy := &obotv1.MCPNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-b"},
		Spec: obotv1.MCPNetworkPolicySpec{
			MCPServerName: "server-b",
			PodSelector:   map[string]string{"app": "server-b"},
			DenyAllEgress: true,
		},
	}

	fp := ToFirewallPolicy(policy, "obot-mcp")
	require.NotNil(t, fp)
	require.Empty(t, fp.Spec.WebGroups)
	require.Len(t, fp.Spec.Rules, 1)
	require.Equal(t, "deny", fp.Spec.Rules[0].Action)
}

func TestTranslateDefaultWildcardDomain(t *testing.T) {
	policy := &obotv1.MCPNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-c"},
		Spec: obotv1.MCPNetworkPolicySpec{
			MCPServerName: "server-c",
			PodSelector:   map[string]string{"app": "server-c"},
		},
	}

	fp := ToFirewallPolicy(policy, "obot-mcp")
	require.NotNil(t, fp)
	require.Len(t, fp.Spec.WebGroups, 1)
	require.Equal(t, []string{"*"}, fp.Spec.WebGroups[0].Domains)
	require.Len(t, fp.Spec.Rules, 2)
	require.Equal(t, "allow-approved-egress", fp.Spec.Rules[0].Name)
	require.Equal(t, "deny-all-external", fp.Spec.Rules[1].Name)
}

func TestProducedObjectShape(t *testing.T) {
	policy := &obotv1.MCPNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-d"},
		Spec: obotv1.MCPNetworkPolicySpec{
			MCPServerName: "server-d",
			PodSelector:   map[string]string{"app": "server-d"},
			EgressDomains: []string{"api.example.com"},
		},
	}

	fp := ToFirewallPolicy(policy, "obot-mcp")
	require.NotNil(t, fp)
	require.Equal(t, aviatrixv1alpha1.SchemeGroupVersion.String(), fp.APIVersion)
	require.Equal(t, "FirewallPolicy", fp.Kind)
	require.Equal(t, "any-destination", fp.Spec.Rules[0].DestinationSmartGroups[0].Name)
}

func TestMCPNetworkPolicyNameFromFirewallPolicyName(t *testing.T) {
	policyName, ok := MCPNetworkPolicyNameFromFirewallPolicyName("obot-policy-a-fw")
	require.True(t, ok)
	require.Equal(t, "policy-a", policyName)

	_, ok = MCPNetworkPolicyNameFromFirewallPolicyName("unmanaged")
	require.False(t, ok)

	longName := "policy-" + strings.Repeat("a", 80)
	_, ok = MCPNetworkPolicyNameFromFirewallPolicyName(NameForMCPNetworkPolicy(longName))
	require.False(t, ok)
}
