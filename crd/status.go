package crd

import (
	"context"
	"reflect"

	"github.com/dogmatiq/dyad"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Resource is a DNS-SD resource that has a status.
type Resource interface {
	client.Object

	domain() string
	status() *Status
}

// Status encapsulates the status of a DNS-SD resource.
type Status struct {
	Conditions          []metav1.Condition `json:"conditions,omitempty"`
	ProviderDescription string             `json:"providerDescription,omitempty"`
	Provider            string             `json:"provider,omitempty"`
	Advertiser          map[string]any     `json:"advertiser,omitempty"`
}

// Condition returns the condition with the given type.
func (s *Status) Condition(t string) metav1.Condition {
	for _, c := range s.Conditions {
		if c.Type == t {
			return c
		}
	}

	return metav1.Condition{
		Type:   t,
		Status: metav1.ConditionUnknown,
	}
}

// StatusUpdate is a function that applies a change to a resource's status.
type StatusUpdate func(Resource, *Status)

// UpdateStatus applies a set of updates to the given resource's status.
func UpdateStatus(
	ctx context.Context,
	cli client.Client,
	res Resource,
	updates ...StatusUpdate,
) error {
	before := res.status()
	after := dyad.Clone(before)

	for _, update := range updates {
		update(res, after)
	}

	if reflect.DeepEqual(before, after) {
		return nil
	}

	*before = *after
	if err := cli.Status().Update(ctx, res); err != nil {
		return err
	}

	return nil
}

// MergeCondition returns a StatusUpdate that merges a new Condition into the
// resource's status.
//
// If a Condition with the same type already exists, it is replaced with the new
// Condition, otherwise the new Condition is appended.
func MergeCondition(c metav1.Condition) StatusUpdate {
	return func(res Resource, s *Status) {
		c.ObservedGeneration = res.GetGeneration()
		c.LastTransitionTime = metav1.Now()

		index := slices.IndexFunc(
			s.Conditions,
			func(x metav1.Condition) bool {
				return x.Type == c.Type
			},
		)

		if index == -1 {
			s.Conditions = append(s.Conditions, c)
			return
		}

		x := s.Conditions[index]

		// Only update the LastTransitionTime if the status has actually
		// transitioned.
		if x.Status == c.Status {
			c.LastTransitionTime = x.LastTransitionTime
		}

		s.Conditions[index] = c
	}
}

// UpdateProviderDescription returns a StatusUpdate that sets the
// ProviderDescription field of the resource's status.
func UpdateProviderDescription(desc string) StatusUpdate {
	return func(_ Resource, s *Status) {
		s.ProviderDescription = desc
	}
}

// AssociateProvider returns a StatusUpdate that sets the Provider and
// Advertiser
// fields of the resource's status.
func AssociateProvider(provider string, advertiser map[string]any) StatusUpdate {
	return func(_ Resource, s *Status) {
		s.Provider = provider
		s.Advertiser = advertiser
	}
}

// If returns a StatusUpdate that conditionally applies other updates.
func If(test bool, updates ...StatusUpdate) StatusUpdate {
	return func(res Resource, s *Status) {
		if test {
			for _, m := range updates {
				m(res, s)
			}
		}
	}
}
