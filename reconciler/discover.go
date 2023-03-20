package reconciler

import (
	"context"
	"strings"
	"time"

	"github.com/dogmatiq/proclaim/crd"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) doDiscover(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (time.Duration, error) {
	ttl, discoverable := r.computeDiscoverable(ctx, res)
	return ttl, r.update(
		res,
		crd.MergeCondition(discoverable),
	)
}

func (r *Reconciler) computeDiscoverable(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (time.Duration, metav1.Condition) {
	instances, err := r.Resolver.EnumerateInstances(
		ctx,
		res.Spec.Instance.ServiceType,
		res.Spec.Instance.Domain,
	)
	if err != nil {
		return 0, crd.DiscoveryErrorCondition(err)
	}

	if !slices.ContainsFunc(
		instances,
		func(v string) bool {
			return strings.EqualFold(v, res.Spec.Instance.Name)
		},
	) {
		crd.NegativeBrowseResult(r.Manager, res)
		return 0, crd.NegativeBrowseResultCondition()
	}

	observed, ok, err := r.Resolver.LookupInstance(
		ctx,
		res.Spec.Instance.Name,
		res.Spec.Instance.ServiceType,
		res.Spec.Instance.Domain,
	)
	if err != nil {
		crd.DiscoveryError(r.Manager, res, err)
		return 0, crd.DiscoveryErrorCondition(err)
	}
	if !ok {
		crd.NegativeLookupResult(r.Manager, res)
		return 0, crd.NegativeLookupResultCondition()
	}

	desired := res.Spec.ToDissolve()

	// The TTL of the observed instance may be less than the desired TTL based
	// on how old the DNS server's cache is. So long as the observed TTL does
	// not *exceed* the desired TTL, we consider the records to be in sync.
	if observed.TTL <= desired.TTL {
		desired.TTL = observed.TTL
		if observed.Equal(desired) {
			crd.Discovered(r.Manager, res)
			return observed.TTL, crd.DiscoveredCondition()
		}
	}

	crd.LookupResultOutOfSync(r.Manager, res)
	return observed.TTL, crd.LookupResultOutOfSyncCondition()
}
