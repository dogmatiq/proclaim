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
	// that DNS-SD services are unadvertised when they're underlying resource
	// is deleted.
	FinalizerName = GroupName + "/unadvertise"

	// Version is the version of the API/CRDs.
	Version = "v1"
)

// Status is an enumeration of the possible states of a service instance.
type Status string

const (
	// StatusPending indicates that none of the Proclaim controllers that have
	// reconciled the resource have been configured to advertise on its domain.
	StatusPending Status = "Pending"

	// StatusAdvertising indicates that a controller has identified where to
	// create/update the DNS records and will soon attempt to do so.
	StatusAdvertising Status = "Advertising"

	// StatusAdvertiseError indicates that there was an upstream problem with
	// the provider while attempting to advertise the service instance.
	StatusAdvertiseError Status = "AdvertiseError"

	// StatusAdvertised indicates that the service instance has been advertised
	// successfully.
	StatusAdvertised Status = "Advertised"

	// StatusUnadvertising indicates that a controller has begin to remove
	// the DNS records for the service instance.
	StatusUnadvertising Status = "Unadvertising"

	// StatusUnadvertiseError indicates that there was an upstream problem with
	// the provider while attempting to unadvertise the service instance.
	StatusUnadvertiseError Status = "UnadvertiseError"

	// StatusUnadvertised indicates that the service instance has been
	// unadvertised successfully. This status will rarely be seen as it is set
	// shortly before Kubernetes deletes the resource entirely.
	StatusUnadvertised Status = "Unadvertised"
)

// Discoverability is an enumeration of the possible discoverability outcomes (via
// DNS-SD) for a service instance.
type Discoverability string

const (
	// DiscoverabilityUnknown indicates that the discoverability of the service
	// instance is unknown.
	DiscoverabilityUnknown Discoverability = "Unknown"

	// DiscoverabilityNone indicates that the service instance is not
	// discoverable via DNS-SD.
	DiscoverabilityNone Discoverability = "None"

	// DiscoverabilityPartial indicates that the service instance is
	// discoverable via DNS-SD, but its properties do not match the desired
	// values.
	DiscoverabilityPartial Discoverability = "Partial"

	// DiscoverabilityComplete indicates that the service instance is
	// discoverable via DNS-SD.
	DiscoverabilityComplete Discoverability = "Complete"
)

// DNSSDServiceInstanceSpec is the specification for a service instance.
type DNSSDServiceInstanceSpec struct {
	InstanceName string              `json:"instanceName"`
	ServiceType  string              `json:"serviceType"`
	Domain       string              `json:"domain"`
	TargetHost   string              `json:"targetHost"`
	TargetPort   uint16              `json:"targetPort"`
	Priority     uint16              `json:"priority,omitempty"`
	Weight       uint16              `json:"weight,omitempty"`
	Attributes   []map[string]string `json:"attributes,omitempty"`
	TTL          uint16              `json:"ttl,omitempty"`
}

// DNSSDServiceInstanceStatus contains the status of a service instance.
type DNSSDServiceInstanceStatus struct {
	ProviderDescription string          `json:"providerDescription,omitempty"`
	ProviderID          string          `json:"providerID,omitempty"`
	AdvertiserID        string          `json:"advertiserID,omitempty"`
	AdvertiseGeneration int64           `json:"advertiseGeneration,omitempty"`
	Status              Status          `json:"status,omitempty"`
	Discoverability     Discoverability `json:"discoverability,omitempty"`
	AdvertisedAt        metav1.Time     `json:"advertisedAt,omitempty"`
	ProbedAt            metav1.Time     `json:"probedAt,omitempty"`
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
