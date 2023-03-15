package dnsimplex

import (
	"sort"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"golang.org/x/exp/slices"
)

// RecordHasAttributes returns true if the attributes of r are equivalent to the
// values in a.
func RecordHasAttributes(
	r dnsimple.ZoneRecord,
	a dnsimple.ZoneRecordAttributes,
) bool {
	if r.Type != a.Type {
		return false
	}

	if r.Name != *a.Name {
		return false
	}

	if r.Content != a.Content {
		return false
	}

	if r.TTL != a.TTL {
		return false
	}

	if r.Priority != a.Priority {
		return false
	}

	recRegions := slices.Clone(r.Regions)
	sort.Strings(recRegions)

	attrRegions := slices.Clone(a.Regions)
	sort.Strings(attrRegions)

	// Treat an empty slice as equivalent to "global", as this is what is
	// returned by the API when a record is created via a plan that does not
	// support regions.
	if len(attrRegions) == 0 {
		attrRegions = []string{"global"}
	}

	return slices.Equal(recRegions, attrRegions)
}
