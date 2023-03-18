package crd

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ConditionTypeReady is a condition that indicates whether or not
	// the service instance is discoverable via the DNS system.
	ConditionTypeReady = "Ready"
)

// ReadyConditionDiscoverable returns a condition indicating that the DNS-SD discovery
// results match the desired state.
func ReadyConditionDiscoverable() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "Discoverable",
		Message: "DNS-SD discovery results match the desired state.",
	}
}

// ReadyConditionNotDiscoverable returns a condition indicating that the DNS-SD
// discovery results did not find the instance.
func ReadyConditionNotDiscoverable() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  "NotDiscoverable",
		Message: "DNS-SD discovery results indicate that this instance does not exist.",
	}
}

// ReadyConditionOutOfSync returns a condition indicating that the DNS-SD
// discovery results found the instance but it differs from the desired state.
func ReadyConditionOutOfSync() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  "OutOfSync",
		Message: "DNS-SD discovery results indicate that instance exists, but it differs from the desired state.",
	}
}

// ReadyConditionError returns a condition indicating that the DNS-SD discovery
// failed with the given error.
func ReadyConditionError(err error) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  "Error",
		Message: fmt.Sprintf("DNS-SD discovery failed: %s", err),
	}
}

// Condition merges a new Condition into the resource's status.
// )
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
