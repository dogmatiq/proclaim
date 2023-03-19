package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/proclaim/crd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// advertise adds/updates DNS records to ensure the given service instance is
// advertised.
//
// It returns true on success. A non-nil error indicates context cancelation or
// a problem interacting with Kubernetes itself.
func (r *Reconciler) advertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (reconcile.Result, error) {
	// If the resource does not have a finalizer, add one. This ensures that
	// we are notified on deletion and have an opportunity to unadvertise the
	// service.
	if controllerutil.AddFinalizer(res, crd.FinalizerName) {
		if err := r.Client.Update(ctx, res); err != nil {
			return reconcile.Result{}, fmt.Errorf("unable to add finalizer: %w", err)
		}
	}

	if shouldAdvertise(res) {
		if err := r.doAdvertise(ctx, res); err != nil {
			return reconcile.Result{}, err
		}
	}

	if shouldDiscover(res) {
		ttl, err := r.doDiscover(ctx, res)
		return shouldRequeue(res, ttl), err
	}

	return shouldRequeue(res, 0), nil
}

func (r *Reconciler) doAdvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) error {
	// Get the advertiser used for this service instance's domain, looking it up
	// by domain if necessary.
	a, ok, err := r.getOrAssociateAdvertiser(ctx, res)
	if !ok || err != nil {
		return err
	}

	// Update the DNS records to reflect the service instance's existence.
	cs, err := a.Advertise(ctx, instanceFromSpec(res.Spec))

	advertised := res.Condition(crd.ConditionTypeAdvertised)

	if err != nil {
		crd.ProviderError(
			r.Manager,
			res,
			res.Status.ProviderID,
			res.Status.ProviderDescription,
			err,
		)
		advertised = crd.AdvertiseErrorCondition(err)
	} else if cs.IsEmpty() {
		crd.DNSRecordsVerified(r.Manager, res)
		if advertised.Status != metav1.ConditionTrue {
			advertised = crd.DNSRecordsObservedCondition()
		}
	} else if cs.IsCreate() {
		crd.DNSRecordsCreated(r.Manager, res)
		advertised = crd.DNSRecordsCreatedCondition()
	} else {
		crd.DNSRecordsUpdated(r.Manager, res)
		advertised = crd.DNSRecordsUpdatedCondition()
	}

	return r.update(
		res,
		crd.MergeCondition(advertised),
	)
}

func shouldAdvertise(res *crd.DNSSDServiceInstance) bool {
	a := res.Condition(crd.ConditionTypeAdvertised)
	d := res.Condition(crd.ConditionTypeDiscoverable)

	if a.Status != metav1.ConditionTrue {
		return true
	}

	if a.ObservedGeneration < res.Generation {
		return true
	}

	if d.Status == metav1.ConditionTrue {
		return false
	}

	return false
}

func shouldDiscover(res *crd.DNSSDServiceInstance) bool {
	a := res.Condition(crd.ConditionTypeAdvertised)

	if a.Status != metav1.ConditionTrue {
		return false
	}

	if a.ObservedGeneration < res.Generation {
		return false
	}

	return true
}

func shouldRequeue(res *crd.DNSSDServiceInstance, discoveredTTL time.Duration) reconcile.Result {
	a := res.Condition(crd.ConditionTypeAdvertised)
	d := res.Condition(crd.ConditionTypeDiscoverable)

	if a.Status != metav1.ConditionTrue {
		return reconcile.Result{Requeue: true}
	}

	if a.ObservedGeneration < res.Generation {
		return reconcile.Result{Requeue: true}
	}

	if d.Status == metav1.ConditionTrue {
		return reconcile.Result{}
	}

	if discoveredTTL == 0 {
		// We have no TTL information of "out of sync" DNS records, so we use
		// the TTL from the specification.
		//
		// HACK: This doesn't really have anything to do with the TTL, we're
		// just using it as a (hopefully) reasonable indicator of how long we
		// should wait before re-trying. It would be better if the provider
		// could give us retry intervals based on the zone's SOA record (e.g.
		// negative cache times) and/or API rate limiting.
		return reconcile.Result{
			RequeueAfter: res.Spec.Instance.TTL.Duration,
		}
	}

	// Otherwise, we wait long enough for the mismatching discovered DNS records
	// to expire (plus a small buffer).
	return reconcile.Result{
		RequeueAfter: discoveredTTL + (1 * time.Second),
	}
}
