package crd

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ConditionTypeDiscoverable is a condition that indicates whether or not the
// service instance is discoverable via the DNS system.
const ConditionTypeDiscoverable = "Discoverable"

// Discovered records an event indicating that the service instance was
// discovered via DNS-SD.
func Discovered(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-dnssd").
		Event(
			res,
			"Normal",
			"Discovered",
			"instance discovered",
		)
}

// DiscoveredCondition returns a condition indicating that the DNS-SD
// discovery results match the advertised DNS records.
func DiscoveredCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionTrue,
		Reason:  "Discovered",
		Message: "DNS-SD browse and lookup results match the advertised DNS records",
	}
}

// NegativeBrowseResult records an event indicating that the service instance
// was not discoverable via DNS-SD.
func NegativeBrowseResult(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-dnssd").
		Event(
			res,
			"Warning",
			"NegativeBrowseResult",
			"instance not discovered",
		)
}

// NegativeBrowseResultCondition returns a condition indicating that the
// instance was not present in the result of a DNS-SD browse operation.
func NegativeBrowseResultCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionFalse,
		Reason:  "NegativeBrowseResult",
		Message: "DNS-SD browse could not find this instance",
	}
}

// NegativeLookupResult records an event indicating that the service instance
// was not discoverable via DNS-SD.
func NegativeLookupResult(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-dnssd").
		Event(
			res,
			"Warning",
			"NegativeLookupResult",
			"instance not discovered",
		)
}

// NegativeLookupResultCondition returns a condition indicating that the
// instance could not be found by a DNS-SD lookup operation.
func NegativeLookupResultCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionFalse,
		Reason:  "NegativeLookupResult",
		Message: "DNS-SD lookup could not find this instance",
	}
}

// LookupResultOutOfSync records an event indicating that the service instance
// was discovered via DNS-SD, but the result did not match the advertised DNS
// records.
func LookupResultOutOfSync(m manager.Manager, res *DNSSDServiceInstance, diff string) {
	m.
		GetEventRecorderFor("proclaim-dnssd").
		Event(
			res,
			"Warning",
			"LookupResultOutOfSync",
			"instance discovered with incorrect (potentially cached) values: "+diff,
		)
}

// LookupResultOutOfSyncCondition returns a condition indicating that the
// instance was found by a DNS-SD lookup operation, but the result did not match
// the advertised DNS records.
func LookupResultOutOfSyncCondition(diff string) metav1.Condition {
	return metav1.Condition{
		Type:   ConditionTypeDiscoverable,
		Status: metav1.ConditionFalse,
		Reason: "LookupResultOutOfSync",
		Message: fmt.Sprintf(
			"DNS-SD lookup result does not match the advertised DNS records: %s",
			diff,
		),
	}
}

// DiscoveryError records an event indicating that an error occurred while
// performing DNS-SD discovery.
func DiscoveryError(
	m manager.Manager,
	res *DNSSDServiceInstance,
	err error,
) {
	m.
		GetEventRecorderFor("proclaim-dnssd").
		Eventf(
			res,
			"Warning",
			"DiscoveryError",
			"%s",
			err.Error(),
		)
}

// DiscoveryErrorCondition returns a condition indicating that the DNS-SD
// discovery failed with the given error.
func DiscoveryErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDiscoverable,
		Status:  metav1.ConditionFalse,
		Reason:  "Error",
		Message: err.Error(),
	}
}
