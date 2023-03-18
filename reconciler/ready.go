package reconciler

import (
	"context"
	"strings"
	"time"

	"github.com/dogmatiq/proclaim/crd"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) needsAdvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (bool, time.Duration, error) {
	advertised := res.Condition(crd.ConditionTypeAdvertised)
	discoverable, observedTTL := r.computeDiscoverableCondition(ctx, res)

	if err := r.updateStatus(
		res,
		func() {
			res.MergeCondition(discoverable)
		},
	); err != nil {
		return false, 0, err
	}

	// If we haven't advertised the service at this generation of the spec we
	// want to do so immediately.
	//
	// This is necessary even if the current records appear to match the spec.
	// We may be observing cached records that will become wrong in the future.
	// We also can't observe the actual TTL of records if we're querying
	// non-authoritative nameservers, which we almost certainly are, by design.
	if advertised.Status != metav1.ConditionTrue || advertised.ObservedGeneration < res.Generation {
		return true, 0, nil
	}

	// If we've advertised the records at this generation of the spec, and the
	// DNS-SD results match the spec, we're done.
	if discoverable.Status == metav1.ConditionTrue {
		return false, 0, nil
	}

	// We found a service instance, but if it doesn't match the spec we
	// reconcile again after the existing records have (hopefully) expired.
	if observedTTL >= 0 {
		const expiryBuffer = 5 * time.Second
		return true, observedTTL + expiryBuffer, nil
	}

	// We didn't find an existing service, so we're not waiting for any existing
	// DNS records to expire.
	//
	// Instead, we wait for the "desired" TTL to pass. We do so under the
	// assumption that the domain's SOA record and any other relevant TTLs are
	// configured appropriately to be used with the TTL from the spec.
	//
	// TODO: This is a bit of a hack, and we should probably have the provider
	// supply information about its SOA records and rate-limiting.
	elapsed := time.Since(advertised.LastTransitionTime.Time)
	return true, res.Spec.Instance.TTL.Duration - elapsed, nil
}

func (r *Reconciler) computeDiscoverableCondition(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (metav1.Condition, time.Duration) {
	instances, err := r.Resolver.EnumerateInstances(
		ctx,
		res.Spec.Instance.ServiceType,
		res.Spec.Instance.Domain,
	)
	if err != nil {
		return crd.DiscoverableConditionError(err), -1
	}

	if !slices.ContainsFunc(
		instances,
		func(v string) bool {
			return strings.EqualFold(v, res.Spec.Instance.Name)
		},
	) {
		return crd.DiscoverableConditionNegativeBrowseResult(), -1
	}

	observed, ok, err := r.Resolver.LookupInstance(
		ctx,
		res.Spec.Instance.Name,
		res.Spec.Instance.ServiceType,
		res.Spec.Instance.Domain,
	)
	if err != nil {
		return crd.DiscoverableConditionError(err), -1
	}
	if !ok {
		return crd.DiscoverableConditionNegativeLookupResult(), -1
	}

	desired := instanceFromSpec(res.Spec)

	// The TTL of the observed instance may be less than the desired TTL based
	// on how old the DNS server's cache is. So long as the observed TTL does
	// not *exceed* the desired TTL, we consider the records to be in sync.
	if observed.TTL <= desired.TTL {
		desired.TTL = observed.TTL
		if observed.Equal(desired) {
			return crd.DiscoverableConditionDiscoverable(), observed.TTL
		}
	}

	return crd.DiscoverableConditionLookupResultOutOfSync(), observed.TTL
}
