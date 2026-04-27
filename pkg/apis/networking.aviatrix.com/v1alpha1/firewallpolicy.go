package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FirewallPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []FirewallPolicy `json:"items"`
}
