package reconciler

import (
	"context"
	"fmt"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// unadvertise removes DNS records to ensure the given service instance is no
// longer advertised.
//
// It returns true on success. A non-nil error indicates context cancelation or
// a problem interacting with Kubernetes itself.
func (r *Reconciler) unadvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
	inst dnssd.ServiceInstance,
) (bool, error) {
	if !controllerutil.ContainsFinalizer(res, crd.FinalizerName) {
		// The proclaim finalizer has already been removed so we've already
		// unadvertised this service instance (if it was even necessary).
		return true, nil
	}

	if len(res.Status.ProviderPayload) != 0 {
		a, ok, err := r.getAdvertiser(ctx, res)
		if !ok || err != nil {
			// The associated provider is not known to this reconciler.
			return true, err
		}

		if res.Status.Status != crd.StatusUnadvertiseError {
			if err := r.setStatus(ctx, res, crd.StatusUnadvertising); err != nil {
				return false, err
			}
		}

		result, err := a.Unadvertise(ctx, inst)

		switch result {
		case provider.UnadvertiseError:
			r.EventRecorder.Eventf(
				res,
				"Warning",
				"Error",
				"%s: %s",
				res.Status.ProviderDescription,
				err.Error(),
			)

			if err := r.setStatus(ctx, res, crd.StatusUnadvertiseError); err != nil {
				return false, err
			}

			return false, ctx.Err()

		case provider.InstanceNotAdvertised:
			// The service instance was not advertised, so we don't need to
			// push another event.

		case provider.UnadvertisedExistingInstance:
			r.EventRecorder.Eventf(
				res,
				"Normal",
				"Unadvertised",
				"service instance unadvertised successfully",
			)
		}
	}

	if err := r.setStatus(ctx, res, crd.StatusUnadvertised); err != nil {
		return false, err
	}

	controllerutil.RemoveFinalizer(res, crd.FinalizerName)
	if err := r.Client.Update(ctx, res); err != nil {
		return false, fmt.Errorf("unable to remove finalizer: %w", err)
	}

	return true, nil
}
