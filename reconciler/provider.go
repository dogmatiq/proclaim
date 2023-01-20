package reconciler

import (
	"context"
	"fmt"
	"strings"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"golang.org/x/exp/slices"
)

func (r *Reconciler) advertiserForResource(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	if res.Status.Provider == "" {
		return r.assignAdvertiser(ctx, res)
	}

	a, ok := r.getAssignedAdvertiser(ctx, res)
	return a, ok, nil
}

func (r *Reconciler) assignAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	var providers []string

	hasProvideError := false
	for _, p := range r.Providers {
		providers = append(providers, p.ID())

		a, ok, err := p.AdvertiserByDomain(ctx, res.Spec.Domain)
		if err != nil {
			hasProvideError = true
			r.EventRecorder.AnnotatedEventf(
				res,
				map[string]string{
					"provider": p.ID(),
				},
				"Warning",
				"ProviderError",
				"unable to check whether the %q provider can advertise on %q: %s",
				p.ID(),
				res.Spec.Domain,
				err.Error(),
			)

			continue
		}

		if !ok {
			continue
		}

		res.Status.Provider = p.ID()
		res.Status.Advertiser = a.ID()

		if err := r.Client.Status().Update(ctx, res); err != nil {
			return nil, false, fmt.Errorf("unable to update resource status: %w", err)
		}

		r.EventRecorder.AnnotatedEventf(
			res,
			map[string]string{
				"provider":   p.ID(),
				"advertiser": a.ID(),
			},
			"Normal",
			"ProviderAssigned",
			"assigned the service instance to the %q provider",
			p.ID(),
		)

		return a, true, nil
	}

	// If any of the providers returned an error, we do not want to emit a
	// DomainIgnored event because haven't checked that all of the providers
	// definitevely do not advertise for this domain.
	if hasProvideError {
		return nil, false, nil
	}

	slices.Sort(providers)

	r.EventRecorder.AnnotatedEventf(
		res,
		map[string]string{
			"providers": strings.Join(providers, ","),
		},
		"Warning",
		"Ignored",
		"none of the configured providers can advertise on %q",
		res.Spec.Domain,
	)

	return nil, false, nil
}

func (r *Reconciler) getAssignedAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool) {
	for _, p := range r.Providers {
		if p.ID() != res.Status.Provider {
			continue
		}

		a, err := p.AdvertiserByID(ctx, res.Status.Advertiser)
		if err != nil {
			r.EventRecorder.AnnotatedEventf(
				res,
				map[string]string{
					"provider":   p.ID(),
					"advertiser": res.Status.Advertiser,
				},
				"Warning",
				"ProviderError",
				"unable to obtain advertiser %q from the %q provider: %s",
				res.Status.Advertiser,
				p.ID(),
				err.Error(),
			)

			return nil, false
		}

		return a, true
	}

	// This reconciler does not know about the provider that was assigned to the
	// resource. This is likely because the resource is managed by some other
	// instance of Proclaim.
	return nil, false
}
