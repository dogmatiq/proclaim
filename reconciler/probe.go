package reconciler

import (
	"context"
	"strings"

	"github.com/dogmatiq/proclaim/crd"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) probe(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) error {
	disc := r.discover(ctx, res)

	return r.updateStatus(
		ctx,
		res,
		func(s *crd.DNSSDServiceInstanceStatus) {
			s.Discoverability = disc
			s.ProbedAt = metav1.Now()
		},
	)
}

func (r *Reconciler) discover(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) crd.Discoverability {
	discovered, exists, err := r.Resolver.LookupInstance(
		ctx,
		res.Spec.Instance.Name,
		res.Spec.Instance.ServiceType,
		res.Spec.Instance.Domain,
	)
	if err != nil {
		return crd.DiscoverabilityUnknown
	}

	synced := false
	if exists {
		desired := instanceFromSpec(res.Spec)

		// The TTL of the discovered record may be less than the desired TTL
		// depending on when the DNS server cached the record. So long as the
		// discovered TTL does not *exceed* the desired TTL, we consider the
		// records to be in sync.
		if discovered.TTL <= desired.TTL {
			desired.TTL = discovered.TTL
		}

		synced = discovered.Equal(desired)
	}

	names, err := r.Resolver.EnumerateInstances(
		ctx,
		res.Spec.Instance.ServiceType,
		res.Spec.Instance.Domain,
	)
	if err != nil {
		return crd.DiscoverabilityUnknown
	}

	enumerable := slices.ContainsFunc(
		names,
		func(v string) bool {
			return strings.EqualFold(v, res.Spec.Instance.Name)
		},
	)

	if enumerable && synced {
		return crd.DiscoverabilityComplete
	}

	if enumerable || exists {
		return crd.DiscoverabilityPartial
	}

	return crd.DiscoverabilityNone
}
