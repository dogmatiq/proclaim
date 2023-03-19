package crd

import (
	"context"
	"reflect"
	"time"

	"github.com/dogmatiq/dyad"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DNSSDServiceInstanceStatus contains the status of a service instance.
type DNSSDServiceInstanceStatus struct {
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
	LastReconciled metav1.Time        `json:"lastReconciled,omitempty"`

	ProviderDescription string `json:"providerDescription,omitempty"`
	ProviderID          string `json:"providerID,omitempty"`
	AdvertiserID        string `json:"advertiserID,omitempty"`
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

// UpdateStatus applies the given StatusUpdates to the given resource.
func UpdateStatus(
	ctx context.Context,
	cli client.Client,
	res *DNSSDServiceInstance,
	updates ...StatusUpdate,
) error {
	clone := dyad.Clone(res)

	for _, update := range updates {
		update(clone)
	}

	if reflect.DeepEqual(clone.Status, res.Status) {
		return nil
	}

	if err := cli.Status().Update(ctx, clone); err != nil {
		return err
	}

	*res = *clone

	return nil
}

// StatusUpdate is a function that updates a resource's status in some way.
type StatusUpdate func(*DNSSDServiceInstance)

// MergeCondition is an StatusUpdate that merges a new Condition into the
// resource's status.
//
// If a Condition with the same type already exists, it is replaced with the new
// Condition, otherwise the new Condition is appended.
func MergeCondition(c metav1.Condition) StatusUpdate {
	return func(res *DNSSDServiceInstance) {
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
}

// UpdateLastReconciled is an StatusUpdate that sets the LastReconciled field of
// the resource's status.
func UpdateLastReconciled(t time.Time) StatusUpdate {
	return func(res *DNSSDServiceInstance) {
		res.Status.LastReconciled = metav1.NewTime(t)
	}
}

// UpdateProviderDescription is an StatusUpdate that sets the
// ProviderDescription field of the resource's status.
func UpdateProviderDescription(desc string) StatusUpdate {
	return func(res *DNSSDServiceInstance) {
		res.Status.ProviderDescription = desc
	}
}

// AssociateProvider is an StatusUpdate that sets the ProviderID and
// AdvertiserID fields of the resource's status.
func AssociateProvider(providerID, advertiserID string) StatusUpdate {
	return func(res *DNSSDServiceInstance) {
		res.Status.ProviderID = providerID
		res.Status.AdvertiserID = advertiserID
	}
}

// If is an StatusUpdate that conditionally applies other StatusUpdates.
func If(test bool, updates ...StatusUpdate) StatusUpdate {
	return func(res *DNSSDServiceInstance) {
		if test {
			for _, update := range updates {
				update(res)
			}
		}
	}
}
