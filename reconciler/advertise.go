package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	probeInterval     = 10 * time.Second
	advertiseInterval = 1 * time.Minute
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

	// If we've advertised this instance at its current generation *and* those
	// records are reflected in actual DNS queries, there's nothing left to do.
	if res.Status.AdvertiseGeneration == res.Generation {
		if err := r.probe(ctx, res); err != nil {
			return reconcile.Result{}, err
		}

		if res.Status.Discoverability == crd.DiscoverabilityComplete {
			return reconcile.Result{}, nil
		}

		if time.Since(res.Status.AdvertisedAt.Time) < advertiseInterval {
			return reconcile.Result{
				RequeueAfter: probeInterval,
			}, nil
		}
	}

	// Otherwise, we have to make sure the correct DNS records are in place.
	// Maybe the service has never been advertised, or maybe the records have
	// been manipulated by some other process or person.

	// Get the advertiser used for this service instance's domain, looking it up
	// by domain if necessary.
	a, ok, err := r.getOrAssociateAdvertiser(ctx, res)
	if !ok || err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	// Update the DNS records to reflect the service instance's existence.
	result, err := a.Advertise(ctx, instanceFromSpec(res.Spec))

	// Record an event about the result of the advertisement.
	switch result {
	case provider.AdvertiseError:
		r.EventRecorder.Eventf(
			res,
			"Warning",
			"Error",
			"%s: %s",
			res.Status.ProviderDescription,
			err,
		)

		if err := r.updateStatus(
			ctx,
			res,
			func(s *crd.DNSSDServiceInstanceStatus) {
				s.Status = crd.StatusAdvertiseError
			},
		); err != nil {
			return reconcile.Result{}, err
		}

		// Requeue for retry.
		return reconcile.Result{Requeue: true}, ctx.Err()

	case provider.InstanceAlreadyAdvertised:
		// The service instance is already advertised, so we don't need to do
		// push another event.

	case provider.AdvertisedNewInstance:
		r.EventRecorder.Eventf(
			res,
			"Normal",
			"Advertised",
			"advertised new service instance",
		)

	case provider.UpdatedExistingInstance:
		r.EventRecorder.Eventf(
			res,
			"Normal",
			"Updated",
			"updating existing service instance",
		)
	}

	if err := r.updateStatus(
		ctx,
		res,
		func(s *crd.DNSSDServiceInstanceStatus) {
			s.AdvertiseGeneration = res.Generation
			s.Status = crd.StatusAdvertised
			s.AdvertisedAt = metav1.Now()
		},
	); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{
		RequeueAfter: probeInterval,
	}, nil
}
