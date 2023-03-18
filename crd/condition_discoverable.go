package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ConditionTypeDiscoverable is a condition that indicates whether or not
	// the service instance is discoverable via the DNS system.
	ConditionTypeDiscoverable = "Discoverable"
)

// DiscoverableConditionDiscoverable returns a condition indicating that the
// DNS-SD discovery results match the advertised DNS records.
func DiscoverableConditionDiscoverable() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionTrue,
		Reason:  "Discoverable",
		Message: "DNS-SD browse and lookup results match the advertised DNS records.",
	}
}

// DiscoverableConditionNegativeBrowseResult returns a condition indicating
// that the instance was not present in the result of a DNS-SD browse
// operation.
func DiscoverableConditionNegativeBrowseResult() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionFalse,
		Reason:  "NegativeBrowseResult",
		Message: "DNS-SD browse (aka enumerate) could not find this instance.",
	}
}

// DiscoverableConditionNegativeLookupResult returns a condition indicating that
// the instance could not be found by a DNS-SD lookup operation.
func DiscoverableConditionNegativeLookupResult() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionFalse,
		Reason:  "NegativeLookupResult",
		Message: "DNS-SD lookup could not find this instance.",
	}
}

// DiscoverableConditionLookupResultOutOfSync returns a condition indicating
// that the instance was found by a DNS-SD lookup operation, but the result did
// not match the advertised DNS records.
func DiscoverableConditionLookupResultOutOfSync() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionFalse,
		Reason:  "LookupResultOutOfSync",
		Message: "DNS-SD lookup result does not match the advertised DNS records.",
	}
}

// DiscoverableConditionError returns a condition indicating that the DNS-SD
// discovery failed with the given error.
func DiscoverableConditionError(err error) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionFalse,
		Reason:  "Error",
		Message: err.Error(),
	}
}
