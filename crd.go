package proclaim

import (
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/mohae/deepcopy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	// FinalizerName is the name of the finalizer that is added to instances
	// so that they can be cleaned up when they are deleted.
	finalizerName = groupName + "/finalizer"
	groupName     = "dns-sd.proclaim.dogmatiq.io"
)

// InstanceSpec is the specification for a service instance.
type InstanceSpec struct {
	Name       string              `json:"name"`
	Service    string              `json:"service"`
	Domain     string              `json:"domain"`
	TargetHost string              `json:"targetHost"`
	TargetPort uint16              `json:"targetPort"`
	Priority   uint16              `json:"priority,omitempty"`
	Weight     uint16              `json:"weight,omitempty"`
	Attributes []map[string]string `json:"attributes,omitempty"`
	TTL        uint16              `json:"ttl,omitempty"`
}

// Instance is a resource that represents a DNS-SD service instance.
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstanceSpec `json:"spec,omitempty"`
}

// DeepCopyObject returns a deep clone of i.
func (i *Instance) DeepCopyObject() runtime.Object {
	if i == nil {
		return nil
	}

	return deepcopy.Copy(i).(*Instance)
}

// InstanceList is a list of DNS-SD service instances.
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Instance `json:"items"`
}

// DeepCopyObject returns a deep clone of l.
func (l *InstanceList) DeepCopyObject() runtime.Object {
	if l == nil {
		return nil
	}

	return deepcopy.Copy(l).(*InstanceList)
}

// SchemeBuilder is the scheme builder for the CRD.
var SchemeBuilder = &scheme.Builder{
	GroupVersion: schema.GroupVersion{
		Group:   groupName,
		Version: "v1alpha1",
	},
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}

// newInstanceFromSpec returns a dnssd.Instance from a specification.
func newInstanceFromSpec(spec InstanceSpec) dnssd.ServiceInstance {
	result := dnssd.ServiceInstance{
		Instance:    spec.Name,
		ServiceType: spec.Service,
		Domain:      spec.Domain,
		TargetHost:  spec.TargetHost,
		TargetPort:  spec.TargetPort,
		Priority:    spec.Priority,
		Weight:      spec.Weight,
		TTL:         time.Duration(spec.TTL) * time.Second,
	}

	if result.TTL == 0 {
		result.TTL = 60 * time.Second
	}

	for _, src := range spec.Attributes {
		var dst dnssd.Attributes

		for k, v := range src {
			if v == "" {
				dst.SetFlag(k)
			} else {
				dst.Set(k, []byte(v))
			}
		}

		result.Attributes = append(result.Attributes, dst)
	}

	return result
}
