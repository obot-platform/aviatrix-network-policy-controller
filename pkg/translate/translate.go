package translate

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	aviatrixv1alpha1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/networking.aviatrix.com/v1alpha1"
	obotv1 "github.com/obot-platform/aviatrix-network-policy-controller/pkg/apis/obot.obot.ai/v1"
	"github.com/obot-platform/nah/pkg/name"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LabelManagedBy      = "obot.ai/managed-by"
	LabelSourceName     = "obot.ai/mcp-network-policy-name"
	LabelSourceServer   = "obot.ai/mcp-server-name"
	AnnotationSourceUID = "obot.ai/mcp-network-policy-uid"
	ManagedByValue      = "aviatrix-network-policy-controller"
)

var invalidNameChars = regexp.MustCompile(`[^a-z0-9-]+`)

func ToFirewallPolicy(policy *obotv1.MCPNetworkPolicy, runtimeNamespace string) (*aviatrixv1alpha1.FirewallPolicy, error) {
	if policy == nil {
		return nil, fmt.Errorf("mcpnetworkpolicy is nil")
	}
	if len(policy.Spec.PodSelector) == 0 {
		return nil, fmt.Errorf("mcpnetworkpolicy %s/%s has empty podSelector", policy.Namespace, policy.Name)
	}

	domains := slices.Clone(policy.Spec.EgressDomains)
	if len(domains) == 0 && !policy.Spec.DenyAllEgress {
		domains = []string{"*"}
	}
	slices.Sort(domains)

	fp := &aviatrixv1alpha1.FirewallPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: aviatrixv1alpha1.SchemeGroupVersion.String(),
			Kind:       "FirewallPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      NameForMCPNetworkPolicy(policy.Name),
			Namespace: runtimeNamespace,
			Labels: map[string]string{
				LabelManagedBy:    ManagedByValue,
				LabelSourceName:   policy.Name,
				LabelSourceServer: policy.Spec.MCPServerName,
			},
			Annotations: map[string]string{
				AnnotationSourceUID: string(policy.UID),
			},
		},
		Spec: aviatrixv1alpha1.FirewallPolicySpec{
			SmartGroups: []aviatrixv1alpha1.SmartGroup{
				{
					Name: fmt.Sprintf("obot-%s-pods", policy.Spec.MCPServerName),
					Selectors: []aviatrixv1alpha1.SmartGroupSelector{{
						Type:         "k8s",
						K8sNamespace: runtimeNamespace,
						Tags:         maps.Clone(policy.Spec.PodSelector),
					}},
				},
				{
					Name: "any-destination",
					Selectors: []aviatrixv1alpha1.SmartGroupSelector{{
						CIDR: "0.0.0.0/0",
					}},
				},
			},
		},
	}

	if len(domains) > 0 {
		fp.Spec.WebGroups = []aviatrixv1alpha1.WebGroup{{
			Name:    fmt.Sprintf("obot-%s-approved-domains", policy.Spec.MCPServerName),
			Domains: domains,
		}}
		fp.Spec.Rules = append(fp.Spec.Rules, aviatrixv1alpha1.Rule{
			Name:   "allow-approved-egress",
			Action: "permit",
			Selector: &aviatrixv1alpha1.RuleSelector{
				MatchLabels: maps.Clone(policy.Spec.PodSelector),
			},
			DestinationSmartGroups: []aviatrixv1alpha1.SmartGroupRef{{Name: "any-destination"}},
			WebGroups:              []aviatrixv1alpha1.SmartGroupRef{{Name: fmt.Sprintf("obot-%s-approved-domains", policy.Spec.MCPServerName)}},
			Protocol:               "tcp",
			Port:                   443,
			Logging:                true,
		})
	}

	if policy.Spec.DenyAllEgress || len(domains) > 0 {
		fp.Spec.Rules = append(fp.Spec.Rules, aviatrixv1alpha1.Rule{
			Name:   "deny-all-external",
			Action: "deny",
			Selector: &aviatrixv1alpha1.RuleSelector{
				MatchLabels: maps.Clone(policy.Spec.PodSelector),
			},
			DestinationSmartGroups: []aviatrixv1alpha1.SmartGroupRef{{Name: "any-destination"}},
			Protocol:               "any",
			Logging:                true,
		})
	}

	return fp, nil
}

func NameForMCPNetworkPolicy(policyName string) string {
	return name.SafeConcatName("obot", sanitizeName(policyName), "fw")
}

func MCPNetworkPolicyNameFromFirewallPolicyName(firewallName string) (string, bool) {
	policyName, ok := strings.CutPrefix(firewallName, "obot-")
	if !ok {
		return "", false
	}
	policyName, ok = strings.CutSuffix(policyName, "-fw")
	if !ok || policyName == "" {
		return "", false
	}
	if NameForMCPNetworkPolicy(policyName) != firewallName {
		return "", false
	}
	return policyName, true
}

func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = invalidNameChars.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	return name
}
