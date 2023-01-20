package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler manipulates DNS records to match the state of a
// crd.DNSSDServiceInstance.
type Reconciler struct {
	Client        client.Client
	EventRecorder record.EventRecorder
	Providers     []provider.Provider
}

// Reconcile performs a full reconciliation for the object referred to by the
// Request, which must be a crd.DNSSDServiceInstance.
func (r *Reconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	res := &crd.DNSSDServiceInstance{}
	if err := r.Client.Get(ctx, req.NamespacedName, res); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	op := r.advertise
	if !res.ObjectMeta.DeletionTimestamp.IsZero() {
		op = r.unadvertise
	}

	ok, err := op(
		ctx,
		res,
		newInstanceFromSpec(res.Spec),
	)

	return reconcile.Result{Requeue: !ok}, err
}

func (r *Reconciler) advertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
	inst dnssd.ServiceInstance,
) (bool, error) {
	if !controllerutil.ContainsFinalizer(res, crd.FinalizerName) {
		controllerutil.AddFinalizer(res, crd.FinalizerName)

		if err := r.Client.Update(ctx, res); err != nil {
			return false, fmt.Errorf("unable to add finalizer: %w", err)
		}

		// The update to the finalizer will cause the object to be requeued, at
		// which point we actually attempt to advertise it.
		return true, nil
	}

	a, ok, err := r.advertiserForResource(ctx, res)
	if !ok || err != nil {
		return false, err
	}

	if err := a.Advertise(ctx, inst); err != nil {
		r.EventRecorder.AnnotatedEventf(
			res,
			map[string]string{
				"provider":   res.Status.Provider,
				"advertiser": res.Status.Advertiser,
			},
			"Warning",
			"ProviderError",
			"unable to advertise the service instance using the %q provider: %w",
			res.Status.Provider,
			err.Error(),
		)

		return false, nil
	}

	r.EventRecorder.AnnotatedEventf(
		res,
		map[string]string{
			"provider":   res.Status.Provider,
			"advertiser": res.Status.Advertiser,
		},
		"Normal",
		"Advertised",
		"service instance advertised using the %q provider",
		res.Status.Provider,
	)

	return true, nil
}

func (r *Reconciler) unadvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
	inst dnssd.ServiceInstance,
) (bool, error) {
	if !controllerutil.ContainsFinalizer(res, crd.FinalizerName) {
		// The proclaim finalizer has already been removed so we've already
		// unadvertised, if it was necessary.
		return true, nil
	}

	if res.Status.Provider == "" {
		// The resource was never assigned to any provider, so there's nothing
		// to unadvertise.
		return true, nil
	}

	adv, ok := r.getAssignedAdvertiser(ctx, res)
	if !ok {
		// The assigned provider is not known to this reconciler.
		return true, nil
	}

	if err := adv.Unadvertise(ctx, inst); err != nil {
		r.EventRecorder.AnnotatedEventf(
			res,
			map[string]string{
				"provider":   res.Status.Provider,
				"advertiser": res.Status.Advertiser,
			},
			"Warning",
			"ProviderError",
			"unable to un-advertise service instance using the %q provider: %w",
			res.Status.Provider,
			err.Error(),
		)

		return false, nil
	}

	r.EventRecorder.AnnotatedEventf(
		res,
		map[string]string{
			"provider":   res.Status.Provider,
			"advertiser": res.Status.Advertiser,
		},
		"Normal",
		"Unadvertised",
		"service instance un-advertised using the %q provider",
		res.Status.Provider,
	)

	controllerutil.RemoveFinalizer(res, crd.FinalizerName)
	if err := r.Client.Update(ctx, res); err != nil {
		return false, fmt.Errorf("unable to remove finalizer: %w", err)
	}

	return true, nil
}

// newInstanceFromSpec returns a dnssd.Instance from a specification.
func newInstanceFromSpec(spec crd.DNSSDServiceInstanceSpec) dnssd.ServiceInstance {
	result := dnssd.ServiceInstance{
		Instance:    spec.Name,
		ServiceType: spec.Service,
		Domain:      spec.Domain,
		TargetHost:  spec.TargetHost,
		TargetPort:  spec.TargetPort,
		Priority:    spec.Priority,
		Weight:      spec.Weight,
		TTL:         time.Duration(spec.TTL) * time.Second,
	}

	if result.TTL == 0 {
		result.TTL = 60 * time.Second
	}

	for _, src := range spec.Attributes {
		var dst dnssd.Attributes

		for k, v := range src {
			if v == "" {
				dst.SetFlag(k)
			} else {
				dst.Set(k, []byte(v))
			}
		}

		result.Attributes = append(result.Attributes, dst)
	}

	return result
}
