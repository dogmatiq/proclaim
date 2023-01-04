package proclaim

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles DNS records with a Instance CRD.
type Reconciler struct {
	Client  client.Client
	Drivers []Driver
}

// Reconcile performs a full reconciliation for the object referred to by the
// Request, which must be a crd.Instance.
func (r *Reconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	obj := &Instance{}
	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	logger := log.FromContext(ctx).V(1).WithValues(
		"instance-name", obj.Spec.Name,
		"service-type", obj.Spec.Service,
		"domain", obj.Spec.Domain,
	)

	driver, adv, err := r.advertiserForDomain(ctx, logger, obj.Spec.Domain)
	if err != nil {
		return reconcile.Result{}, err
	}

	logger = logger.WithValues(
		"driver", reflect.TypeOf(driver).String(),
	)

	op := r.advertise
	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		op = r.unadvertise
	}

	err = op(
		ctx,
		adv,
		logger,
		obj,
		newInstanceFromSpec(obj.Spec),
	)

	// TODO: update status instead of returning an error

	return reconcile.Result{}, err
}

func (r *Reconciler) advertiserForDomain(
	ctx context.Context,
	logger logr.Logger,
	domain string,
) (Driver, Advertiser, error) {
	for _, d := range r.Drivers {
		adv, ok, err := d.AdvertiserForDomain(ctx, logger, domain)
		if err != nil {
			logger.Error(
				err,
				"driver failed when attempting to obtain an advertiser",
				"driver", reflect.TypeOf(d).String(),
			)

			continue
		}

		if ok {
			return d, adv, nil
		}
	}

	err := errors.New("could not find advertiser")

	logger.Error(
		err,
		"none of the configured drivers could provide an advertiser for this domain",
	)

	return nil, nil, err
}

func (r *Reconciler) advertise(
	ctx context.Context,
	adv Advertiser,
	logger logr.Logger,
	obj *Instance,
	inst dnssd.ServiceInstance,
) error {
	// Only advertise the instance if the finalizer has already been added.
	if controllerutil.ContainsFinalizer(obj, finalizerName) {
		return adv.Advertise(ctx, logger, inst)
	}

	// Otherwise, add the finalizer, which is itself and update and will cause
	// the controller to be called again.
	controllerutil.AddFinalizer(obj, finalizerName)
	if err := r.Client.Update(ctx, obj); err != nil {
		return fmt.Errorf("unable to add finalizer: %w", err)
	}

	logger.Info(
		"added finalizer",
		"finalizer", finalizerName,
	)

	return nil
}

func (r *Reconciler) unadvertise(
	ctx context.Context,
	adv Advertiser,
	logger logr.Logger,
	obj *Instance,
	inst dnssd.ServiceInstance,
) error {
	if err := adv.Unadvertise(ctx, logger, inst); err != nil {
		return err
	}

	if controllerutil.ContainsFinalizer(obj, finalizerName) {
		controllerutil.RemoveFinalizer(obj, finalizerName)
		if err := r.Client.Update(ctx, obj); err != nil {
			return fmt.Errorf("unable to remove finalizer: %w", err)
		}

		logger.Info(
			"removed finalizer",
			"finalizer", finalizerName,
		)
	}

	return nil
}
