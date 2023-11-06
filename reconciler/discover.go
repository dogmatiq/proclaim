package reconciler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
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

	if drift, ok := compare(observed, desired); !ok {
		crd.LookupResultOutOfSync(r.Manager, res, drift)
		return observed.TTL, crd.LookupResultOutOfSyncCondition(drift)
	}

	d := res.Condition(crd.ConditionTypeDiscoverable)
	if d.Status != metav1.ConditionTrue {
		crd.Discovered(r.Manager, res)
	}
	return observed.TTL, crd.DiscoveredCondition()
}

// compare returns a (very) brief human-readable description of the differences
// between the observed and desired service instance records.
func compare(observed, desired dnssd.ServiceInstance) (string, bool) {
	if observed.TargetHost != desired.TargetHost {
		return fmt.Sprintf("host %q != %q", observed.TargetHost, desired.TargetHost), false
	}

	if observed.TargetPort != desired.TargetPort {
		return fmt.Sprintf("port %d != %d", observed.TargetPort, desired.TargetPort), false
	}

	if observed.Priority != desired.Priority {
		return fmt.Sprintf("priority %d != %d", observed.Priority, desired.Priority), false
	}

	if observed.Weight != desired.Weight {
		return fmt.Sprintf("weight %d != %d", observed.Weight, desired.Weight), false
	}

	if !dnssd.AttributeCollectionsEqual(observed.Attributes, desired.Attributes) {
		return "attributes", false
	}

	// The TTL of the observed instance may be less than the desired TTL based
	// on how old the DNS server's cache is. So long as the observed TTL does
	// not *exceed* the desired TTL, we consider the records to be in sync.
	if observed.TTL > desired.TTL {
		return fmt.Sprintf("ttl %d > %d", observed.TTL, desired.TTL), false
	}

	return "", true
}
