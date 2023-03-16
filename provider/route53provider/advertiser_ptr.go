package route53provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/dissolve/dnssd"
	"golang.org/x/exp/slices"
)

// ptrTTL is the TTL of PTR records that enumerate service instances.
//
// Normally we'd use each service's TTL for its respective PTR record, but with
// Route 53 the only way to return an unlimited number of PTR records with the
// same name is to put them in the same "record set", which means they all share
// a TTL.
const ptrTTL = 30 * time.Second

func (a *advertiser) findPTR(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (types.ResourceRecordSet, bool, error) {
	return a.findResourceRecordSet(
		ctx,
		serviceName(inst),
		types.RRTypePtr,
	)
}

func (a *advertiser) syncPTR(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *types.ChangeBatch,
) error {
	desired := types.ResourceRecordSet{
		SetIdentifier: marshalGeneration(0),
		Weight:        aws.Int64(0),
		Type:          types.RRTypePtr,
		Name:          serviceName(inst),
		TTL:           aws.Int64(int64(ptrTTL.Seconds())),
		ResourceRecords: convertRecords(
			dnssd.NewPTRRecord(inst),
		),
	}

	current, ok, err := a.findPTR(ctx, inst)
	if err != nil {
		return err
	}

	if !ok {
		cs.Changes = append(
			cs.Changes,
			types.Change{
				Action:            types.ChangeActionCreate,
				ResourceRecordSet: &desired,
			},
		)

		return nil
	}

	if indexOf(current, inst) != -1 {
		return nil
	}

	gen, err := unmarshalGeneration(current.SetIdentifier)
	if err != nil {
		return err
	}

	desired.SetIdentifier = marshalGeneration(gen + 1)
	desired.ResourceRecords = append(desired.ResourceRecords, current.ResourceRecords...)

	cs.Changes = append(
		cs.Changes,
		types.Change{
			Action:            types.ChangeActionCreate,
			ResourceRecordSet: &desired,
		},
		types.Change{
			Action:            types.ChangeActionDelete,
			ResourceRecordSet: &current,
		},
	)

	return nil
}

func (a *advertiser) deletePTR(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *types.ChangeBatch,
) error {
	current, ok, err := a.findPTR(ctx, inst)
	if !ok || err != nil {
		return err
	}

	index := indexOf(current, inst)
	if index == -1 {
		return nil
	}

	gen, err := unmarshalGeneration(current.SetIdentifier)
	if err != nil {
		return err
	}

	cs.Changes = append(
		cs.Changes,
		types.Change{
			Action:            types.ChangeActionDelete,
			ResourceRecordSet: &current,
		},
	)

	desired := types.ResourceRecordSet{
		SetIdentifier: marshalGeneration(gen + 1),
		Weight:        aws.Int64(0),
		Type:          types.RRTypePtr,
		Name:          serviceName(inst),
		TTL:           aws.Int64(int64(ptrTTL.Seconds())),
		ResourceRecords: slices.Delete(
			slices.Clone(current.ResourceRecords),
			index,
			index+1,
		),
	}

	if len(desired.ResourceRecords) != 0 {
		cs.Changes = append(
			cs.Changes,
			types.Change{
				Action:            types.ChangeActionCreate,
				ResourceRecordSet: &desired,
			},
		)
	}

	return nil
}

// indexOf returns the index of the given inst in a PTR resource record set, or
// -1 if it is not present.
func indexOf(set types.ResourceRecordSet, inst dnssd.ServiceInstance) int {
	n := instanceName(inst)
	for i, rec := range set.ResourceRecords {
		if strings.EqualFold(*rec.Value, *n) {
			return i
		}
	}

	return -1
}

const generationPrefix = "dogmatiq/proclaim:generation="

// marshalGeneration returns a string representation of the given generation
// number suitable for being encoded in the SetIdentifier field of a Route 53
// resource record set.
//
// Encoding the generation here allows us to identify resource record sets with
// the same name and type by their generation (version).
func marshalGeneration(n uint64) *string {
	return aws.String(fmt.Sprintf("%s%d", generationPrefix, n))
}

// unmarshalGeneration returns the generation number encoded in the
// SetIdentifier field of a Route 53 resource record set.
func unmarshalGeneration(gen *string) (uint64, error) {
	if gen == nil {
		return 0, errors.New("missing rr-set generation")
	}

	v, ok := strings.CutPrefix(*gen, generationPrefix)
	if !ok {
		return 0, fmt.Errorf("invalid rr-set generation %q: missing prefix", *gen)
	}

	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rr-set generation %q: invalid counter component", *gen)
	}

	return n, nil
}
