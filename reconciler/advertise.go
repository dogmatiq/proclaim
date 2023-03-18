package reconciler

import (
	"context"
	"fmt"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
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

	// Update the DNS records to reflect the service instance's existence.
	result, err := a.Advertise(ctx, instanceFromSpec(res.Spec))

	advertised := res.Condition(crd.ConditionTypeAdvertised)

	// Record an event about the result of the advertisement.
	switch result {
	case provider.AdvertiseError:
		advertised = crd.AdvertisedConditionAdvertiseError(err)
		r.EventRecorder.Eventf(
			res,
			"Warning",
			"Error",
			"%s: %s",
			res.Status.ProviderDescription,
			err,
		)
	case provider.InstanceAlreadyAdvertised:
		if advertised.Status != metav1.ConditionTrue {
			advertised = crd.AdvertisedConditionRecordsCreated()
		}
	case provider.AdvertisedNewInstance:
		advertised = crd.AdvertisedConditionRecordsCreated()
		r.EventRecorder.Eventf(
			res,
			"Normal",
			"Advertised",
			"advertised new service instance",
		)
	case provider.UpdatedExistingInstance:
		advertised = crd.AdvertisedConditionRecordsUpdated()
		r.EventRecorder.Eventf(
			res,
			"Normal",
			"Updated",
			"updating existing service instance",
		)
	}

	discoverable, _ := r.computeDiscoverableCondition(ctx, res)

	if err := r.updateStatus(
		res,
		func() {
			res.MergeCondition(advertised)
			res.MergeCondition(discoverable)

			if result != provider.AdvertiseError {
				res.Status.LastAdvertised = metav1.Now()
			}
		},
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
