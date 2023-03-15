package crd

import (
	"github.com/dogmatiq/dyad"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// GroupName is the API group name used by Proclaim.
	GroupName = "proclaim.dogmatiq.io"

	// FinalizerName is the name of the finalizer used by Proclaim to ensure
	// that DNS-SD services are un-advertised when they're underlying resource
	// is deleted.
	FinalizerName = GroupName + "/unadvertise"

	// Version is the version of the API/CRDs.
	Version = "v1alpha1"
)

// DNSSDServiceInstanceSpec is the specification for a service instance.
type DNSSDServiceInstanceSpec struct {
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

// DNSSDServiceInstanceStatus contains the status of a service instance.
type DNSSDServiceInstanceStatus struct {
	ProviderID          string `json:"providerId,omitempty"`
	ProviderDescription string `json:"providerDescription,omitempty"`
	AdvertiserID        string `json:"advertiserId,omitempty"`
}

// DNSSDServiceInstance is a resource that represents a DNS-SD service instance.
type DNSSDServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSSDServiceInstanceSpec   `json:"spec,omitempty"`
	Status DNSSDServiceInstanceStatus `json:"status,omitempty"`
}

// DeepCopyObject returns a deep clone of i.
func (i *DNSSDServiceInstance) DeepCopyObject() runtime.Object {
	return dyad.Clone(i)
}

// DNSSDServiceInstanceList is a list of DNS-SD service instances.
type DNSSDServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DNSSDServiceInstance `json:"items"`
}

// DeepCopyObject returns a deep clone of l.
func (l *DNSSDServiceInstanceList) DeepCopyObject() runtime.Object {
	return dyad.Clone(l)
}
