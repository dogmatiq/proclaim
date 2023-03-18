package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ConditionTypeAdvertised is a condition that indicates whether or not the
	// service instance has been advertised via a provider.
	ConditionTypeAdvertised = "Advertised"
)

// AdvertisedConditionRecordsCreated returns a condition indicating that the
// instance's DNS records have been created.
func AdvertisedConditionRecordsCreated() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionTrue,
		Reason:  "RecordsCreated",
		Message: "DNS records created successfully.",
	}
}

// AdvertisedConditionRecordsUpdated returns a condition indicating that the
// instance's DNS records have been updated.
func AdvertisedConditionRecordsUpdated() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionTrue,
		Reason:  "RecordsUpdated",
		Message: "DNS records updated successfully.",
	}
}

// AdvertisedConditionRecordsRemoved returns a condition indicating that the
// instance's DNS records have been removed.
func AdvertisedConditionRecordsRemoved() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionFalse,
		Reason:  "RecordsRemoved",
		Message: "DNS records removed successfully.",
	}
}

// AdvertisedConditionAdvertiseError returns a condition indicating that an
// attempt to advertise the instance failed with the given error.
func AdvertisedConditionAdvertiseError(err error) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionUnknown,
		Reason:  "AdvertiseError",
		Message: err.Error(),
	}
}

// AdvertisedConditionUnadvertiseError returns a condition indicating that an
// attempt to unadvertise the instance failed with the given error.
func AdvertisedConditionUnadvertiseError(err error) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdvertised,
		Status:  metav1.ConditionUnknown,
		Reason:  "UnadvertiseError",
		Message: err.Error(),
	}
}
