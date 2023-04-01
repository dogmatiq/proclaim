package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ConditionTypeAdopted is a condition that indicates whether or not the
// service instance has been adopted by a provider.
const ConditionTypeAdopted = "Adopted"

// InstanceAdopted records an event indicating that the service instance was
// adopted by the controller.
func InstanceAdopted(m manager.Manager, res Resource) {
	m.
		GetEventRecorderFor("proclaim-"+res.Status().Provider).
		Eventf(
			res,
			"Normal",
			"InstanceAdopted",
			"%s can advertise on %q",
			res.Status().ProviderDescription,
			res.DissolveName().Domain,
		)
}

// InstanceAdoptedCondition returns a condition indicating that the instance
// has been adopted by a provider.
func InstanceAdoptedCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdopted,
		Status:  metav1.ConditionTrue,
		Reason:  "InstanceAdopted",
		Message: "at least one Proclaim controller has a provider that can advertise on this domain",
	}
}

// InstanceIgnored records an event indicating that the service instance was
// ignored by the controller.
func InstanceIgnored(m manager.Manager, res Resource) {
	m.
		GetEventRecorderFor("proclaim").
		Eventf(
			res,
			"Warning",
			"InstanceIgnored",
			"none of the configured providers can advertise on %q",
			res.DissolveName().Domain,
		)
}

// InstanceIgnoredCondition returns a condition indicating that the instance
// has been ignored by all providers.
func InstanceIgnoredCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeAdopted,
		Status:  metav1.ConditionFalse,
		Reason:  "InstanceIgnored",
		Message: "no running Proclaim controllers have providers that can advertise on this domain",
	}
}
