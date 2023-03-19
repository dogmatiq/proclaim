package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ConditionTypeAdvertised is a condition that indicates whether or not the
// service instance has been advertised via a provider.
const ConditionTypeAdvertised = "Advertised"

// DNSRecordsCreated records an event indicating that new DNS records were
// created.
func DNSRecordsCreated(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-"+res.Status.ProviderID).
		Event(
			res,
			"Normal",
			"RecordsCreated",
			"created new DNS records",
		)
}

// DNSRecordsCreatedCondition returns a condition indicating that the
// instance's DNS records have been created.
func DNSRecordsCreatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionTrue,
		Reason:  "RecordsCreated",
		Message: "created new DNS records",
	}
}

// DNSRecordsUpdated records an event indicating that existing DNS records were
// updated.
func DNSRecordsUpdated(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-"+res.Status.ProviderID).
		Event(
			res,
			"Normal",
			"RecordsUpdated",
			"updated existing DNS records",
		)
}

// DNSRecordsUpdatedCondition returns a condition indicating that the instance's
// DNS records have been updated.
func DNSRecordsUpdatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionTrue,
		Reason:  "RecordsUpdated",
		Message: "updating existing DNS records",
	}
}

// DNSRecordsVerified records an event indicating that existing DNS records were
// verified to match the service instance spec.
func DNSRecordsVerified(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-"+res.Status.ProviderID).
		Event(
			res,
			"Normal",
			"RecordsVerified",
			"verified that existing DNS records have expected values",
		)
}

// DNSRecordsObservedCondition returns a condition indicating that the
// instance's DNS records have been observed to already exist.
func DNSRecordsObservedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionTrue,
		Reason:  "RecordsObserved",
		Message: "found existing DNS records",
	}
}

// DNSRecordsDeleted records an event indicating that existing DNS records were
// deleted.
func DNSRecordsDeleted(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-"+res.Status.ProviderID).
		Event(
			res,
			"Normal",
			"RecordsDeleted",
			"deleted existing DNS records",
		)
}

// DNSRecordsDeletedCondition returns a condition indicating that the instance's
// DNS records have been removed.
func DNSRecordsDeletedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionFalse,
		Reason:  "RecordsDeleted",
		Message: "deleted existing DNS records",
	}
}

// ProviderError records an event indicating that an error occurred while
// interacting with a DNS provider.
func ProviderError(
	m manager.Manager,
	res *DNSSDServiceInstance,
	id, desc string,
	err error,
) {
	m.
		GetEventRecorderFor("proclaim-"+id).
		Eventf(
			res,
			"Warning",
			"ProviderError",
			"%s: %s",
			desc,
			err.Error(),
		)
}

// AdvertiseErrorCondition returns a condition indicating that an attempt to
// advertise the instance failed with the given error.
func AdvertiseErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionUnknown,
		Reason:  "AdvertiseError",
		Message: err.Error(),
	}
}

// UnadvertiseErrorCondition returns a condition indicating that an attempt to
// unadvertise the instance failed with the given error.
func UnadvertiseErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionUnknown,
		Reason:  "UnadvertiseError",
		Message: err.Error(),
	}
}
