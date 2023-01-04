package proclaim

import (
	"context"
	"fmt"
	"sort"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"golang.org/x/exp/slices"
)

func forEach[T any](
	ctx context.Context,
	list func(dnsimple.ListOptions) ([]T, error),
	fn func(T) (bool, error),
) error {
	page := 0
	opt := dnsimple.ListOptions{
		Page: &page,
	}

	for {
		page++
		data, err := list(opt)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return nil
		}

		for _, v := range data {
			ok, err := fn(v)
			if !ok || err != nil {
				return err
			}
		}
	}
}

func recordHasMatchingAttributes(
	rec dnsimple.ZoneRecord,
	attr dnsimple.ZoneRecordAttributes,
) bool {
	if rec.Type != attr.Type {
		panic(fmt.Sprintf("record type mismatch (%s != %s)", rec.Type, attr.Type))
	}

	if rec.Name != *attr.Name {
		panic(fmt.Sprintf("record name mismatch (%s != %s)", rec.Name, *attr.Name))
	}

	if rec.Content != attr.Content {
		return false
	}

	if rec.TTL != attr.TTL {
		return false
	}

	if rec.Priority != attr.Priority {
		return false
	}

	currentRegions := slices.Clone(rec.Regions)
	sort.Strings(currentRegions)

	desiredRegions := slices.Clone(attr.Regions)
	sort.Strings(desiredRegions)

	// Treat an empty slice as equivalent to "global", as this is what is
	// returned by the API when a record is created via a plan that does not
	// support regions.
	if len(desiredRegions) == 0 {
		desiredRegions = []string{"global"}
	}

	return slices.Equal(currentRegions, desiredRegions)
}

// recordName returns the record name to use for advertising a DNS-SD
// service instance.
func recordName(name, service string) *string {
	n := dnssd.EscapeInstance(name) + "." + service
	return dnsimple.String(n)
}
