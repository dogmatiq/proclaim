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
	current, present, err := r.Resolver.LookupInstance(
		ctx,
		res.Spec.InstanceName,
		res.Spec.ServiceType,
		res.Spec.Domain,
	)
	if err != nil {
		return crd.DiscoverabilityUnknown
	}

	synced := false
	if present {
		desired := instanceFromSpec(res.Spec)
		synced = current.Equal(desired)
	}

	names, err := r.Resolver.EnumerateInstances(
		ctx,
		res.Spec.ServiceType,
		res.Spec.Domain,
	)
	if err != nil {
		return crd.DiscoverabilityUnknown
	}

	enumerable := slices.ContainsFunc(
		names,
		func(v string) bool {
			return strings.EqualFold(v, res.Spec.InstanceName)
		},
	)

	if enumerable && synced {
		return crd.DiscoverabilityComplete
	}

	if enumerable || present {
		return crd.DiscoverabilityPartial
	}

	return crd.DiscoverabilityNone
}
