package route53provider

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/miekg/dns"
)

// enumerationRecordTTL is the TTL of PTR records that enumerate service
// instances.
//
// Normally we'd use each service's TTL for its respective PTR record, but with
// Route 53 the only way to return an unlimited number of PTR records with the
// same name is to put them in the same "record set", which means they all share
// a TTL.
const enumerationRecordTTL = 30 * time.Second

func addInstanceToPTRChanges(
	current *route53.ResourceRecordSet,
	instanceName, serviceName *string,
) ([]*route53.Change, error) {
	desired := &route53.ResourceRecordSet{
		SetIdentifier: aws.String("0"),
		Weight:        aws.Int64(0),
		Type:          aws.String("PTR"),
		Name:          serviceName,
		TTL:           aws.Int64(int64(enumerationRecordTTL.Seconds())),
		ResourceRecords: []*route53.ResourceRecord{
			{Value: instanceName},
		},
	}

	if current == nil {
		return []*route53.Change{
			{
				Action:            aws.String(route53.ChangeActionCreate),
				ResourceRecordSet: desired,
			},
		}, nil
	}

	if ptrHasInstance(current, instanceName) {
		return nil, nil
	}

	version, err := strconv.ParseUint(*current.SetIdentifier, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse version: %w", err)
	}

	desired.SetIdentifier = aws.String(strconv.FormatUint(version+1, 10))
	desired.ResourceRecords = append(desired.ResourceRecords, current.ResourceRecords...)

	return []*route53.Change{
		{
			Action:            aws.String(route53.ChangeActionDelete),
			ResourceRecordSet: current,
		},
		{
			Action:            aws.String(route53.ChangeActionCreate),
			ResourceRecordSet: desired,
		},
	}, nil
}

func removeInstanceFromPTRChanges(
	current *route53.ResourceRecordSet,
	instanceName, serviceName *string,
) ([]*route53.Change, error) {
	if current == nil {
		return nil, nil
	}

	if !ptrHasInstance(current, instanceName) {
		return nil, nil
	}

	if len(current.ResourceRecords) == 1 {
		return []*route53.Change{
			{
				Action:            aws.String(route53.ChangeActionDelete),
				ResourceRecordSet: current,
			},
		}, nil
	}

	version, err := strconv.ParseUint(*current.SetIdentifier, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse version: %w", err)
	}

	desired := &route53.ResourceRecordSet{
		SetIdentifier: aws.String(strconv.FormatUint(version+1, 10)),
		Weight:        aws.Int64(0),
		Type:          aws.String("PTR"),
		Name:          serviceName,
		TTL:           aws.Int64(int64(enumerationRecordTTL.Seconds())),
	}

	for _, rec := range current.ResourceRecords {
		if !strings.EqualFold(*rec.Value, *instanceName) {
			desired.ResourceRecords = append(desired.ResourceRecords, rec)
		}
	}

	return []*route53.Change{
		{
			Action:            aws.String(route53.ChangeActionDelete),
			ResourceRecordSet: current,
		},
		{
			Action:            aws.String(route53.ChangeActionCreate),
			ResourceRecordSet: desired,
		},
	}, nil
}

// ptrHasInstance returns true if the given resource record set set contains a
// record with the given value.
func ptrHasInstance(set *route53.ResourceRecordSet, instanceName *string) bool {
	for _, rec := range set.ResourceRecords {
		if strings.EqualFold(*rec.Value, *instanceName) {
			return true
		}
	}

	return false
}

// record is an interface for a DNS record from the miekg/dns package.
type record interface {
	Header() *dns.RR_Header
	String() string
}

// convertRecords creates Route53 records from DNS records.
func convertRecords[R record](records ...R) []*route53.ResourceRecord {
	var result []*route53.ResourceRecord

	for _, rec := range records {
		result = append(
			result,
			&route53.ResourceRecord{
				Value: aws.String(
					strings.TrimPrefix(
						rec.String(),
						rec.Header().String(),
					),
				),
			},
		)
	}

	return result
}
