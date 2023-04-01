package crd

import (
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/dyad"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DNSSDServiceInstanceSubType is a resource that represents a DNS-SD service
// instance's service "sub-type".
type DNSSDServiceInstanceSubType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Stat Status `json:"status,omitempty"`
	Spec struct {
		Instance InstanceName `json:"instance"`
	} `json:"spec,omitempty"`
}

// DeepCopyObject returns a deep clone of i.
func (r *DNSSDServiceInstanceSubType) DeepCopyObject() runtime.Object {
	return dyad.Clone(r)
}

// DissolveName returns a Dissolve dnssd.ServiceInstanceName from a CRD service
// instance.
func (r *DNSSDServiceInstanceSubType) DissolveName() dnssd.ServiceInstanceName {
	return dnssd.ServiceInstanceName{
		Name:        r.Spec.Instance.Name,
		ServiceType: r.Spec.Instance.ServiceType,
		Domain:      r.Spec.Instance.Domain,
	}
}

// Status returns the status of the resource.
func (r *DNSSDServiceInstanceSubType) Status() *Status {
	return &r.Stat
}

// DNSSDServiceInstanceSubTypeList is a list of DNS-SD service instance
// sub-types.
type DNSSDServiceInstanceSubTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DNSSDServiceInstanceSubType `json:"items"`
}

// DeepCopyObject returns a deep clone of l.
func (l *DNSSDServiceInstanceSubTypeList) DeepCopyObject() runtime.Object {
	return dyad.Clone(l)
}

// InstanceName is a DNS-SD service instance name.
type InstanceName struct {
	Name        string `json:"name"`
	ServiceType string `json:"serviceType"`
	Domain      string `json:"domain"`
}

// DissolveName returns a Dissolve dnssd.ServiceInstanceName from a CRD service
// instance name.
func (n InstanceName) DissolveName() dnssd.ServiceInstanceName {
	return dnssd.ServiceInstanceName{
		Name:        n.Name,
		ServiceType: n.ServiceType,
		Domain:      n.Domain,
	}
}
