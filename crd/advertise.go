package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ConditionTypeAdvertised is a condition that indicates whether or not the
// service instance has been advertised via a provider.
const ConditionTypeAdvertised = "Advertised"

// DNSRecordsUpdated records an event indicating that DNS records were created
// or updated.
func DNSRecordsUpdated(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-"+res.Status.Provider).
		Event(
			res,
			"Normal",
			"RecordsUpdated",
			"updated DNS records",
		)
}

// DNSRecordsUpdatedCondition returns a condition indicating that the instance's
// DNS records have been created or updated.
func DNSRecordsUpdatedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionTrue,
		Reason:  "RecordsUpdated",
		Message: "updated DNS records",
	}
}

// DNSRecordsVerified records an event indicating that existing DNS records were
// verified to match the service instance spec.
func DNSRecordsVerified(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-"+res.Status.Provider).
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

// DNSRecordsDeleted records an event indicating that DNS records were deleted.
func DNSRecordsDeleted(m manager.Manager, res *DNSSDServiceInstance) {
	m.
		GetEventRecorderFor("proclaim-"+res.Status.Provider).
		Event(
			res,
			"Normal",
			"RecordsDeleted",
			"deleted DNS records",
		)
}

// DNSRecordsDeletedCondition returns a condition indicating that the instance's
// DNS records have been removed.
func DNSRecordsDeletedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionFalse,
		Reason:  "RecordsDeleted",
		Message: "deleted DNS records",
	}
}

// DNSRecordsDoNotExistCondition returns a condition indicating that the
// instance's DNS records do not exist, either because they never did or they
// have already been removed.
func DNSRecordsDoNotExistCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionFalse,
		Reason:  "RecordsDoNotExist",
		Message: "DNS records do not exist",
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
