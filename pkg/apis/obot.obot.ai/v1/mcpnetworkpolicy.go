package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: "obot.obot.ai", Version: "v1"}
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme        = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &MCPNetworkPolicy{}, &MCPNetworkPolicyList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

type MCPNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MCPNetworkPolicySpec `json:"spec,omitempty"`
}

type MCPNetworkPolicySpec struct {
	MCPServerName string            `json:"mcpServerName,omitempty"`
	PodSelector   map[string]string `json:"podSelector,omitempty"`
	EgressDomains []string          `json:"egressDomains,omitempty"`
	DenyAllEgress bool              `json:"denyAllEgress,omitempty"`
}

type MCPNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MCPNetworkPolicy `json:"items"`
}

func (in *MCPNetworkPolicy) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(MCPNetworkPolicy)
	*out = *in
	out.ObjectMeta = *in.ObjectMeta.DeepCopy()
	if in.Spec.PodSelector != nil {
		out.Spec.PodSelector = make(map[string]string, len(in.Spec.PodSelector))
		for k, v := range in.Spec.PodSelector {
			out.Spec.PodSelector[k] = v
		}
	}
	if in.Spec.EgressDomains != nil {
		out.Spec.EgressDomains = append([]string(nil), in.Spec.EgressDomains...)
	}
	return out
}

func (in *MCPNetworkPolicyList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(MCPNetworkPolicyList)
	*out = *in
	if in.Items != nil {
		out.Items = make([]MCPNetworkPolicy, len(in.Items))
		for i := range in.Items {
			out.Items[i] = *in.Items[i].DeepCopyObject().(*MCPNetworkPolicy)
		}
	}
	return out
}
