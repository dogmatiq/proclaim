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

	// Work out if we need to (re)advertise the instance and when.
	needed, after, err := r.needsAdvertise(ctx, res)
	if !needed || err != nil {
		return reconcile.Result{}, err
	}
	if after > 0 {
		return reconcile.Result{RequeueAfter: after}, nil
	}

	// Get the advertiser used for this service instance's domain, looking it up
	// by domain if necessary.
	a, ok, err := r.getOrAssociateAdvertiser(ctx, res)
	if !ok || err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	advertised := res.Condition(crd.ConditionTypeAdvertised)

	// Update the DNS records to reflect the service instance's existence.
	cs, err := a.Advertise(ctx, instanceFromSpec(res.Spec))

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

	_, _, discoverable := r.discover(ctx, res)

	if err := r.update(
		res,
		crd.MergeCondition(advertised),
		crd.MergeCondition(discoverable),
		crd.If(
			advertised.Status == metav1.ConditionTrue,
			crd.UpdateLastReconciled(time.Now()),
		),
	); err != nil {
		return reconcile.Result{}, err
	}

	if advertised.Status != metav1.ConditionTrue {
		return reconcile.Result{Requeue: true}, nil
	}

	if discoverable.Status != metav1.ConditionTrue {
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}
