package route53provider

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/dissolve/dnssd"
)

func (a *advertiser) findSRV(
	ctx context.Context,
	name dnssd.ServiceInstanceName,
) (types.ResourceRecordSet, bool, error) {
	return a.findResourceRecordSet(
		ctx,
		name.Absolute(),
		types.RRTypeSrv,
	)
}

func (a *advertiser) syncSRV(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *types.ChangeBatch,
) error {
	desired := types.ResourceRecordSet{
		Name: aws.String(inst.Absolute()),
		Type: types.RRTypeSrv,
		TTL:  aws.Int64(int64(inst.TTL.Seconds())),
		ResourceRecords: convertRecords(
			dnssd.NewSRVRecord(inst),
		),
	}

	current, ok, err := a.findSRV(ctx, inst.ServiceInstanceName)
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

	if reflect.DeepEqual(current, desired) {
		return nil
	}

	cs.Changes = append(
		cs.Changes,
		types.Change{
			Action:            types.ChangeActionUpsert,
			ResourceRecordSet: &desired,
		},
	)

	return nil
}

func (a *advertiser) deleteSRV(
	ctx context.Context,
	name dnssd.ServiceInstanceName,
	cs *types.ChangeBatch,
) error {
	current, ok, err := a.findSRV(ctx, name)
	if !ok || err != nil {
		return err
	}

	if ok {
		cs.Changes = append(
			cs.Changes,
			types.Change{
				Action:            types.ChangeActionDelete,
				ResourceRecordSet: &current,
			},
		)
	}

	return nil
}
