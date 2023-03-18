package reconciler

import (
	"context"
	"strings"

	"github.com/dogmatiq/proclaim/crd"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) checkReadyCondition(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) metav1.Condition {
	names, err := r.Resolver.EnumerateInstances(
		ctx,
		res.Spec.Instance.ServiceType,
		res.Spec.Instance.Domain,
	)
	if err != nil {
		return crd.ReadyConditionError(err)
	}

	if !slices.ContainsFunc(
		names,
		func(v string) bool {
			return strings.EqualFold(v, res.Spec.Instance.Name)
		},
	) {
		return crd.ReadyConditionNotDiscoverable()
	}

	observed, ok, err := r.Resolver.LookupInstance(
		ctx,
		res.Spec.Instance.Name,
		res.Spec.Instance.ServiceType,
		res.Spec.Instance.Domain,
	)
	if err != nil {
		return crd.ReadyConditionError(err)
	}
	if !ok {
		return crd.ReadyConditionOutOfSync()
	}

	desired := instanceFromSpec(res.Spec)

	// The TTL of the observed instance may be less than the desired TTL based
	// on how old the DNS server's cache is. So long as the observed TTL does
	// not *exceed* the desired TTL, we consider the records to be in sync.
	if observed.TTL > desired.TTL {
		return crd.ReadyConditionOutOfSync()
	}

	desired.TTL = observed.TTL
	if observed.Equal(desired) {
		return crd.ReadyConditionDiscoverable()
	}

	return crd.ReadyConditionOutOfSync()
}
