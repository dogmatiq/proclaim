package crd

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

// DNSSDServiceInstanceStatus contains the status of a service instance.
type DNSSDServiceInstanceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	ProviderDescription string      `json:"providerDescription,omitempty"`
	ProviderID          string      `json:"providerID,omitempty"`
	AdvertiserID        string      `json:"advertiserID,omitempty"`
	AdvertiseGeneration int64       `json:"advertiseGeneration,omitempty"`
	Status              Status      `json:"status,omitempty"`
	AdvertisedAt        metav1.Time `json:"advertisedAt,omitempty"`
}

// Condition merges a new Condition into the resource's status.
//
// If a Condition with the same type already exists, it is replaced with the new
// Condition, otherwise the new Condition is appended.
func (res *DNSSDServiceInstance) Condition(c metav1.Condition) {
	c.ObservedGeneration = res.Generation
	c.LastTransitionTime = metav1.Now()

	for i, x := range res.Status.Conditions {
		if x.Type == c.Type {
			if x.Status == c.Status {
				c.LastTransitionTime = x.LastTransitionTime
			}
			res.Status.Conditions[i] = c
			return
		}
	}

	res.Status.Conditions = append(res.Status.Conditions, c)
}
