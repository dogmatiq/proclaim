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
	if controllerutil.AddFinalizer(res, crd.FinalizerName) {
		if err := r.Client.Update(ctx, res); err != nil {
			return reconcile.Result{}, fmt.Errorf("unable to add finalizer: %w", err)
		}

		r.Logger.Info(
			"added finalizer",
			"namespace", res.Namespace,
			"name", res.Name,
			"finalizer", crd.FinalizerName,
		)
	}

	if r.shouldAdvertise(res) {
		if err := r.doAdvertise(ctx, res); err != nil {
			return reconcile.Result{}, err
		}
	}

	if r.shouldDiscover(res) {
		ttl, err := r.doDiscover(ctx, res)
		return r.requeueResult(res, ttl), err
	}

	return r.requeueResult(res, 0), nil
}

func (r *Reconciler) doAdvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) error {
	a, ok, err := r.getOrAssociateAdvertiser(ctx, res)
	if !ok || err != nil {
		return err
	}

	changed, err := a.Advertise(ctx, res.Spec.ToDissolve())

	advertised := res.Condition(crd.ConditionTypeAdvertised)

	if err != nil {
		crd.ProviderError(
			r.Manager,
			res,
			res.Status.Provider,
			res.Status.ProviderDescription,
			err,
		)
		advertised = crd.AdvertiseErrorCondition(err)
	} else if changed {
		crd.DNSRecordsUpdated(r.Manager, res)
		advertised = crd.DNSRecordsUpdatedCondition()
	} else {
		crd.DNSRecordsVerified(r.Manager, res)
		if advertised.Status != metav1.ConditionTrue {
			advertised = crd.DNSRecordsObservedCondition()
		}
	}

	return r.update(
		res,
		crd.MergeCondition(advertised),
	)
}

func (r *Reconciler) shouldAdvertise(res *crd.DNSSDServiceInstance) bool {
	a := res.Condition(crd.ConditionTypeAdvertised)
	d := res.Condition(crd.ConditionTypeDiscoverable)

	should := false
	reason := ""

	if a.Status != metav1.ConditionTrue {
		should = true
		reason = "not advertised"
	} else if a.ObservedGeneration < res.Generation {
		should = true
		reason = "resource updated since last advertised"
	} else if d.Status != metav1.ConditionTrue {
		should = true
		reason = "not discoverable"
	} else {
		should = false
		reason = "already discoverable"
	}

	message := "not advertising"
	if should {
		message = "advertising"
	}

	r.Logger.Info(
		message,
		"namespace", res.Namespace,
		"name", res.Name,
		"reason", reason,
	)

	return should
}

func (r *Reconciler) shouldDiscover(res *crd.DNSSDServiceInstance) bool {
	a := res.Condition(crd.ConditionTypeAdvertised)
	d := res.Condition(crd.ConditionTypeDiscoverable)

	should := false
	reason := ""

	if a.Status != metav1.ConditionTrue {
		should = false
		reason = "not advertised"
	} else if a.ObservedGeneration < res.Generation {
		should = false
		reason = "resource updated since last advertised"
	} else if d.Status != metav1.ConditionTrue {
		should = true
		reason = "not discoverable"
	} else {
		should = true
		reason = "drift detection"
	}

	message := "not discovering"
	if should {
		message = "discovering"
	}

	r.Logger.Info(
		message,
		"namespace", res.Namespace,
		"name", res.Name,
		"reason", reason,
	)

	return should
}

func (r *Reconciler) requeueResult(
	res *crd.DNSSDServiceInstance,
	discoveredTTL time.Duration,
) reconcile.Result {
	a := res.Condition(crd.ConditionTypeAdvertised)
	d := res.Condition(crd.ConditionTypeDiscoverable)

	var delay time.Duration
	reason := ""

	if a.Status != metav1.ConditionTrue {
		reason = "not advertised"
	} else if a.ObservedGeneration < res.Generation {
		reason = "resource updated since last advertised"
	} else if d.Status == metav1.ConditionTrue {
		reason = "drift detection"
		delay = 10 * res.Spec.Instance.TTL.Duration
	} else if discoveredTTL == 0 {
		// We have no TTL information from actual DNS records, so we compute
		// something based on the TTL.
		//
		// HACK: This doesn't really have anything to do with the TTL, we're
		// just using it as a (hopefully) reasonable indicator of how long we
		// should wait before re-trying. It would be better if the provider
		// could give us retry intervals based on the zone's SOA record (e.g.
		// negative cache times) and/or API rate limiting.
		delay = res.Spec.Instance.TTL.Duration
		reason = "not discoverable"
	} else {
		// Otherwise, we wait long enough for the mismatching discovered DNS
		// records to expire (plus a small buffer).
		delay = discoveredTTL + (1 * time.Second)
		reason = "records are stale or have drifted"
	}

	r.Logger.Info(
		"re-queuing",
		"namespace", res.Namespace,
		"name", res.Name,
		"reason", reason,
		"next", delay,
	)

	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: delay,
	}
}
