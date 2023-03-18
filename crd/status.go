package crd

import (
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DNSSDServiceInstanceStatus contains the status of a service instance.
type DNSSDServiceInstanceStatus struct {
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
	LastAdvertised metav1.Time        `json:"lastAdvertised,omitempty"`

	ProviderDescription string `json:"providerDescription,omitempty"`
	ProviderID          string `json:"providerID,omitempty"`
	AdvertiserID        string `json:"advertiserID,omitempty"`
}

// MergeCondition merges a new Condition into the resource's status.
//
// If a Condition with the same type already exists, it is replaced with the new
// Condition, otherwise the new Condition is appended.
func (res *DNSSDServiceInstance) MergeCondition(c metav1.Condition) {
	c.ObservedGeneration = res.Generation
	c.LastTransitionTime = metav1.Now()

	index := slices.IndexFunc(
		res.Status.Conditions,
		func(x metav1.Condition) bool {
			return x.Type == c.Type
		},
	)

	if index == -1 {
		res.Status.Conditions = append(res.Status.Conditions, c)
		return
	}

	x := res.Status.Conditions[index]

	// Only update the LastTransitionTime if the status has actually
	// transitioned.
	if x.Status == c.Status {
		c.LastTransitionTime = x.LastTransitionTime
	}

	res.Status.Conditions[index] = c
}

// Condition returns the condition with the given type.
func (res *DNSSDServiceInstance) Condition(t string) metav1.Condition {
	for _, c := range res.Status.Conditions {
		if c.Type == t {
			return c
		}
	}

	return metav1.Condition{
		Type:   t,
		Status: metav1.ConditionUnknown,
	}
}
