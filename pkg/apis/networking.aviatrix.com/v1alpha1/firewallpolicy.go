package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type FirewallPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FirewallPolicySpec `json:"spec,omitempty"`
}

type FirewallPolicySpec struct {
	SmartGroups []SmartGroup `json:"smartGroups,omitempty"`
	WebGroups   []WebGroup   `json:"webGroups,omitempty"`
	Rules       []Rule       `json:"rules,omitempty"`
}

type SmartGroup struct {
	Name      string               `json:"name,omitempty"`
	Selectors []SmartGroupSelector `json:"selectors,omitempty"`
}

type SmartGroupSelector struct {
	Type         string            `json:"type,omitempty"`
	K8sNamespace string            `json:"k8sNamespace,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	CIDR         string            `json:"cidr,omitempty"`
}

type WebGroup struct {
	Name    string   `json:"name,omitempty"`
	Domains []string `json:"domains,omitempty"`
}

type Rule struct {
	Name                   string          `json:"name,omitempty"`
	Action                 string          `json:"action,omitempty"`
	Selector               *RuleSelector   `json:"selector,omitempty"`
	DestinationSmartGroups []SmartGroupRef `json:"destinationSmartGroups,omitempty"`
	WebGroups              []SmartGroupRef `json:"webGroups,omitempty"`
	Protocol               string          `json:"protocol,omitempty"`
	Port                   int32           `json:"port,omitempty"`
	Logging                bool            `json:"logging,omitempty"`
}

type RuleSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
	Service     string            `json:"service,omitempty"`
}

type SmartGroupRef struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

type FirewallPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []FirewallPolicy `json:"items"`
}

// TODO(g-linville): auto-generate deepcopy functions

func (in *FirewallPolicy) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(FirewallPolicy)
	*out = *in
	out.ObjectMeta = *in.ObjectMeta.DeepCopy()
	out.Spec = *in.Spec.DeepCopy()
	return out
}

func (in *FirewallPolicyList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(FirewallPolicyList)
	*out = *in
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		out.Items = make([]FirewallPolicy, len(in.Items))
		for i := range in.Items {
			out.Items[i] = *in.Items[i].DeepCopy()
		}
	}
	return out
}

func (in *FirewallPolicy) DeepCopy() *FirewallPolicy {
	if in == nil {
		return nil
	}
	out := new(FirewallPolicy)
	*out = *in
	out.ObjectMeta = *in.ObjectMeta.DeepCopy()
	out.Spec = *in.Spec.DeepCopy()
	return out
}

func (in *FirewallPolicySpec) DeepCopy() *FirewallPolicySpec {
	if in == nil {
		return nil
	}
	out := new(FirewallPolicySpec)
	*out = *in
	if in.SmartGroups != nil {
		out.SmartGroups = make([]SmartGroup, len(in.SmartGroups))
		for i := range in.SmartGroups {
			out.SmartGroups[i] = *in.SmartGroups[i].DeepCopy()
		}
	}
	if in.WebGroups != nil {
		out.WebGroups = make([]WebGroup, len(in.WebGroups))
		for i := range in.WebGroups {
			out.WebGroups[i] = in.WebGroups[i]
			if in.WebGroups[i].Domains != nil {
				out.WebGroups[i].Domains = append([]string(nil), in.WebGroups[i].Domains...)
			}
		}
	}
	if in.Rules != nil {
		out.Rules = make([]Rule, len(in.Rules))
		for i := range in.Rules {
			out.Rules[i] = *in.Rules[i].DeepCopy()
		}
	}
	return out
}

func (in *SmartGroup) DeepCopy() *SmartGroup {
	if in == nil {
		return nil
	}
	out := new(SmartGroup)
	*out = *in
	if in.Selectors != nil {
		out.Selectors = make([]SmartGroupSelector, len(in.Selectors))
		for i := range in.Selectors {
			out.Selectors[i] = *in.Selectors[i].DeepCopy()
		}
	}
	return out
}

func (in *SmartGroupSelector) DeepCopy() *SmartGroupSelector {
	if in == nil {
		return nil
	}
	out := new(SmartGroupSelector)
	*out = *in
	if in.Tags != nil {
		out.Tags = make(map[string]string, len(in.Tags))
		for k, v := range in.Tags {
			out.Tags[k] = v
		}
	}
	return out
}

func (in *Rule) DeepCopy() *Rule {
	if in == nil {
		return nil
	}
	out := new(Rule)
	*out = *in
	if in.Selector != nil {
		out.Selector = in.Selector.DeepCopy()
	}
	if in.DestinationSmartGroups != nil {
		out.DestinationSmartGroups = make([]SmartGroupRef, len(in.DestinationSmartGroups))
		copy(out.DestinationSmartGroups, in.DestinationSmartGroups)
	}
	if in.WebGroups != nil {
		out.WebGroups = make([]SmartGroupRef, len(in.WebGroups))
		copy(out.WebGroups, in.WebGroups)
	}
	return out
}

func (in *RuleSelector) DeepCopy() *RuleSelector {
	if in == nil {
		return nil
	}
	out := new(RuleSelector)
	*out = *in
	if in.MatchLabels != nil {
		out.MatchLabels = make(map[string]string, len(in.MatchLabels))
		for k, v := range in.MatchLabels {
			out.MatchLabels[k] = v
		}
	}
	return out
}
