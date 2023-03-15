package reconciler

import (
	"context"
	"fmt"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// advertise adds/updates DNS records to ensure the given service instance is
// advertised.
//
// It returns true on success. A non-nil error indicates context cancelation or
// a problem interacting with Kubernetes itself.
func (r *Reconciler) advertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
	inst dnssd.ServiceInstance,
) (bool, error) {
	// If the resource does not have a finalizer, add one. This ensures that
	// we are notified on deletion and have an opportunity to unadvertise the
	// service.
	if !controllerutil.ContainsFinalizer(res, crd.FinalizerName) {
		controllerutil.AddFinalizer(res, crd.FinalizerName)

		if err := r.Client.Update(ctx, res); err != nil {
			return false, fmt.Errorf("unable to add finalizer: %w", err)
		}

		// Adding the finalizer will cause the object to be requeued, so rather
		// than advertising it now we will do it on the next pass.
		return true, nil
	}

	// Get the advertiser used for this service instance's domain, looking it up
	// by domain if necessary.
	a, ok, err := r.getOrAssociateAdvertiser(ctx, res)
	if !ok || err != nil {
		return false, err
	}

	// Update the DNS records to reflect the service instance's existence.
	result, err := a.Advertise(ctx, inst)

	// Record an event about the result of the advertisement.
	switch result {
	case provider.AdvertiseError:
		r.EventRecorder.Eventf(
			res,
			"Warning",
			"Error",
			"%s: %w",
			res.Status.ProviderID,
			err,
		)
		return false, ctx.Err()

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

	return true, nil
}
