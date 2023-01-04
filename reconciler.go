package proclaim

import (
	"context"
	"fmt"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles DNS records with a Instance CRD.
type Reconciler struct {
	Client        client.Client
	EventRecorder record.EventRecorder
	Drivers       []Driver
}

// Reconcile performs a full reconciliation for the object referred to by the
// Request, which must be a crd.Instance.
func (r *Reconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	obj := &DNSSDServiceInstance{}
	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	logger := log.FromContext(ctx).V(1).WithValues(
		"instance-name", obj.Spec.Name,
		"service-type", obj.Spec.Service,
		"domain", obj.Spec.Domain,
	)

	driver, adv, ok, err := r.advertiser(ctx, logger, obj)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !ok {
		r.EventRecorder.Eventf(
			obj,
			"Warn",
			"DNSSDAdvertiserUnavailable",
			"none of the configured drivers can advertise DNS-SD services on %q",
			obj.Spec.Domain,
		)

		return reconcile.Result{}, nil
	}

	logger = logger.WithValues(
		"driver", driver.Name(),
	)

	if obj.Status.Driver != driver.Name() {
		obj.Status.Driver = driver.Name()

		if err := r.Client.Status().Update(
			ctx,
			obj,
		); err != nil {
			return reconcile.Result{}, err
		}
	}

	op := r.advertise
	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		op = r.unadvertise
	}

	ok, err = op(
		ctx,
		adv,
		logger,
		obj,
		newInstanceFromSpec(obj.Spec),
	)
	return reconcile.Result{
		Requeue: !ok,
	}, err
}

func (r *Reconciler) advertiser(
	ctx context.Context,
	logger logr.Logger,
	obj *DNSSDServiceInstance,
) (Driver, Advertiser, bool, error) {
	for _, d := range r.Drivers {
		adv, ok, err := d.AdvertiserForDomain(ctx, logger, obj.Spec.Domain)
		if err != nil {
			return nil, nil, false, fmt.Errorf("%s: %w", d.Name(), err)
		}

		if ok {
			return d, adv, true, nil
		}
	}

	return nil, nil, false, nil
}

func (r *Reconciler) advertise(
	ctx context.Context,
	adv Advertiser,
	logger logr.Logger,
	obj *DNSSDServiceInstance,
	inst dnssd.ServiceInstance,
) (bool, error) {
	if !controllerutil.ContainsFinalizer(obj, finalizerName) {
		controllerutil.AddFinalizer(obj, finalizerName)
		return true, r.Client.Update(ctx, obj)
	}

	if err := adv.Advertise(ctx, logger, inst); err != nil {
		r.EventRecorder.Eventf(
			obj,
			"Warning",
			"DNSSDAdvertiseFailed",
			err.Error(),
		)
		return false, nil
	}

	r.EventRecorder.Eventf(
		obj,
		"Warning",
		"DNSSDAdvertiseSucceeded",
		"DNS-SD service instance has been successfully advertised",
	)

	return true, nil
}

func (r *Reconciler) unadvertise(
	ctx context.Context,
	adv Advertiser,
	logger logr.Logger,
	obj *DNSSDServiceInstance,
	inst dnssd.ServiceInstance,
) (bool, error) {
	if err := adv.Unadvertise(ctx, logger, inst); err != nil {
		r.EventRecorder.Eventf(
			obj,
			"Warning",
			"DNSSDUnadvertiseFailed",
			err.Error(),
		)
		return false, nil
	}

	r.EventRecorder.Eventf(
		obj,
		"Warning",
		"DNSSDUnadvertiseSucceeded",
		"DNS-SD service instance has been successfully un-advertised",
	)

	if controllerutil.ContainsFinalizer(obj, finalizerName) {
		controllerutil.RemoveFinalizer(obj, finalizerName)
		return true, r.Client.Update(ctx, obj)
	}

	return true, nil
}
