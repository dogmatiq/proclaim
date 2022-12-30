package proclaim

import (
	"golang.org/x/exp/maps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

type InstanceSpec struct {
	Name       string            `json:"name"`
	Service    string            `json:"service"`
	Domain     string            `json:"domain"`
	Host       string            `json:"host"`
	Port       uint16            `json:"port"`
	Priority   uint16            `json:"priority,omitempty"`
	Weight     uint16            `json:"weight,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (s InstanceSpec) DeepCopy() InstanceSpec {
	s.Attributes = maps.Clone(s.Attributes)
	return s
}

type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstanceSpec `json:"spec,omitempty"`
}

// func (i *DNSSDInstance) ValidateCreate() error {
// 	return nil
// }

// func (i *DNSSDInstance) ValidateUpdate(old runtime.Object) error {
// 	return nil
// }

// func (i *DNSSDInstance) ValidateDelete() error {
// 	return nil
// }

func (i *Instance) DeepCopy() *Instance {
	if i == nil {
		return nil
	}

	clone := *i
	i.ObjectMeta.DeepCopyInto(&clone.ObjectMeta)
	clone.Spec = i.Spec.DeepCopy()

	return &clone
}

func (i *Instance) DeepCopyObject() runtime.Object {
	if i == nil {
		return nil
	}

	return i.DeepCopy()
}

type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Instance `json:"items"`
}

func (l *InstanceList) DeepCopy() *InstanceList {
	if l == nil {
		return nil
	}

	clone := *l
	l.ListMeta.DeepCopyInto(&clone.ListMeta)

	for i, inst := range l.Items {
		clone.Items[i] = *inst.DeepCopy()
	}

	return &clone
}

func (l *InstanceList) DeepCopyObject() runtime.Object {
	if l == nil {
		return nil
	}

	return l.DeepCopy()
}

var SchemeBuilder = &scheme.Builder{
	GroupVersion: schema.GroupVersion{
		Group:   "dns-sd.proclaim.dogmatiq.io",
		Version: "v1alpha1",
	},
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
