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
	groupName     = "proclaim.dogmatiq.io"
	finalizerName = groupName
)

// Spec is the specification for a service instance.
type Spec struct {
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

// DNSSDServiceInstance is a resource that represents a DNS-SD service instance.
type DNSSDServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec Spec `json:"spec,omitempty"`
}

// DeepCopyObject returns a deep clone of i.
func (i *DNSSDServiceInstance) DeepCopyObject() runtime.Object {
	if i == nil {
		return nil
	}

	return deepcopy.Copy(i).(*DNSSDServiceInstance)
}

// DNSSDServiceInstanceList is a list of DNS-SD service instances.
type DNSSDServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DNSSDServiceInstance `json:"items"`
}

// DeepCopyObject returns a deep clone of l.
func (l *DNSSDServiceInstanceList) DeepCopyObject() runtime.Object {
	if l == nil {
		return nil
	}

	return deepcopy.Copy(l).(*DNSSDServiceInstanceList)
}

// SchemeBuilder is the scheme builder for the CRD.
var SchemeBuilder = &scheme.Builder{
	GroupVersion: schema.GroupVersion{
		Group:   groupName,
		Version: "v1alpha1",
	},
}

func init() {
	SchemeBuilder.Register(
		&DNSSDServiceInstance{},
		&DNSSDServiceInstanceList{},
	)
}

// newInstanceFromSpec returns a dnssd.Instance from a specification.
func newInstanceFromSpec(spec Spec) dnssd.ServiceInstance {
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
