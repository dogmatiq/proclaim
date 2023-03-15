package route53provider

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/dissolve/dnssd"
)

func (a *advertiser) findTXT(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (types.ResourceRecordSet, bool, error) {
	return a.findResourceRecordSet(
		ctx,
		instanceName(inst),
		types.RRTypeTxt,
	)
}

func (a *advertiser) syncTXT(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *types.ChangeBatch,
) error {
	desired := types.ResourceRecordSet{
		Name: instanceName(inst),
		Type: types.RRTypeTxt,
		TTL:  aws.Int64(int64(inst.TTL.Seconds())),
		ResourceRecords: convertRecords(
			dnssd.NewTXTRecords(inst)...,
		),
	}

	current, ok, err := a.findTXT(ctx, inst)
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

func (a *advertiser) deleteTXT(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *types.ChangeBatch,
) error {
	current, ok, err := a.findTXT(ctx, inst)
	if err != nil {
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
