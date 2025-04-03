package dnsimplex

import (
	"context"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
)

// Find calls fn for each value returned by list until fn returns true.
//
// It returns the result of fn.
func Find[T, V any](
	ctx context.Context,
	list func(dnsimple.ListOptions) (*dnsimple.Pagination, []T, error),
	fn func(T) (V, bool, error),
) (result V, ok bool, err error) {
	return result, ok, Each(
		ctx,
		list,
		func(v T) (bool, error) {
			result, ok, err = fn(v)
			return !ok, err
		},
	)
}

// Each calls fn for each value returned by list.
//
// list is called once for each page of results.
//
// If fn returns false, the iteration stops.
func Each[T any](
	_ context.Context,
	list func(dnsimple.ListOptions) (*dnsimple.Pagination, []T, error),
	fn func(T) (bool, error),
) error {
	page := 0
	opt := dnsimple.ListOptions{
		Page: &page,
	}

	for {
		page++
		p, data, err := list(opt)
		if err != nil {
			return err
		}

		for _, v := range data {
			ok, err := fn(v)
			if !ok || err != nil {
				return err
			}
		}

		if p == nil || page >= p.TotalPages {
			return nil
		}
	}
}
