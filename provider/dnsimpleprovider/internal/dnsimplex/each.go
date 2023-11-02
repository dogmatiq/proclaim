package dnsimplex

import (
	"context"

	"github.com/dnsimple/dnsimple-go/dnsimple"
)

// All returns a slice of all the values returned by list.
//
// list is called once for each page of results.
func All[T any](
	ctx context.Context,
	list func(dnsimple.ListOptions) (*dnsimple.Pagination, []T, error),
) ([]T, error) {
	var result []T

	err := Each(
		ctx,
		list,
		func(v T) (bool, error) {
			result = append(result, v)
			return true, nil
		},
	)

	return result, err
}

// One returns the first value returned by list.
func One[T any](
	ctx context.Context,
	list func(dnsimple.ListOptions) (*dnsimple.Pagination, []T, error),
) (T, bool, error) {
	var result T
	var ok bool

	err := Each(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []T, error) {
			opts.PerPage = dnsimple.Int(1)
			return list(opts)
		},
		func(v T) (bool, error) {
			result = v
			ok = true
			return false, nil
		},
	)

	return result, ok, err
}

// First returns the first value returned by list for which pred returns true.
//
// list is called once for each page of results.
func First[T any](
	ctx context.Context,
	list func(dnsimple.ListOptions) (*dnsimple.Pagination, []T, error),
	pred func(T) bool,
) (T, bool, error) {
	var result T
	var ok bool

	err := Each(
		ctx,
		list,
		func(v T) (bool, error) {
			if pred(v) {
				result = v
				ok = true
				return false, nil
			}

			return true, nil
		},
	)

	return result, ok, err
}

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
