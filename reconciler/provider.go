package reconciler

import (
	"context"
	"fmt"
	"strings"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"golang.org/x/exp/slices"
)

// advertiserForResource returns the advertiser used to advertise/un-advertise
// the given DNS-SD service instance.
func (r *Reconciler) advertiserForResource(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	if res.Status.Provider == "" {
		return r.assignAdvertiser(ctx, res)
	}

	return r.getAssignedAdvertiser(ctx, res)
}

// assignAdvertiser finds the appropriate advertiser for the given DNS-SD
// service instance from all available providers and assigns it to the resource.
func (r *Reconciler) assignAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	hasProviderError := false

	for _, p := range r.Providers {
		a, ok, err := p.AdvertiserByDomain(ctx, res.Spec.Domain)
		if err != nil {
			hasProviderError = true

			r.EventRecorder.AnnotatedEventf(
				res,
				map[string]string{
					"provider": p.ID(),
				},
				"Warning",
				"ProviderError",
				"%s: %s",
				p.ID(),
				err.Error(),
			)

			if ctx.Err() != nil {
				return nil, false, ctx.Err()
			}

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
			"assigned the %q provider to manage this service instance",
			p.ID(),
		)

		return a, true, nil
	}

	// If any of the providers returned an error, we do not want to emit an
	// Ignored event because we don't know that all of the providers
	// definitevely do not advertise for this domain.
	if hasProviderError {
		return nil, false, nil
	}

	var providers []string
	for _, p := range r.Providers {
		providers = append(providers, p.ID())
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
) (provider.Advertiser, bool, error) {
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
				"%s: %s",
				p.ID(),
				err.Error(),
			)

			return nil, false, ctx.Err()
		}

		return a, true, nil
	}

	// This reconciler does not know about the provider that was assigned to the
	// resource. This is likely because the resource is managed by some other
	// instance of Proclaim.
	return nil, false, nil
}
