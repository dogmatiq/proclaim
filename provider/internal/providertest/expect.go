package providertest

import (
	"context"
	"reflect"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/onsi/gomega"
	"golang.org/x/exp/slices"
)

func expectInstanceToExist(
	ctx context.Context,
	res *dnssd.UnicastResolver,
	expect dnssd.ServiceInstance,
) {
	for {
		actual, ok, err := res.LookupInstance(
			ctx,
			expect.Instance,
			expect.ServiceType,
			expect.Domain,
		)
		gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())

		if ok {
			gomega.ExpectWithOffset(1, actual).To(gomega.Equal(expect))
			return
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func expectInstanceToConverge(
	ctx context.Context,
	res *dnssd.UnicastResolver,
	expect dnssd.ServiceInstance,
) {
	for {
		actual, ok, err := res.LookupInstance(
			ctx,
			expect.Instance,
			expect.ServiceType,
			expect.Domain,
		)
		gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())

		if ok && reflect.DeepEqual(actual, expect) {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func expectInstanceNotToExist(
	ctx context.Context,
	res *dnssd.UnicastResolver,
	expect dnssd.ServiceInstance,
) {
	for {
		_, ok, err := res.LookupInstance(
			ctx,
			expect.Instance,
			expect.ServiceType,
			expect.Domain,
		)
		gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())

		if !ok {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func expectInstanceListToConverge(
	ctx context.Context,
	res *dnssd.UnicastResolver,
	service, domain string,
	expect ...dnssd.ServiceInstance,
) {
	var names []string
	for _, inst := range expect {
		names = append(names, inst.Instance)
	}

	slices.Sort(names)

	for {
		instances, err := res.EnumerateInstances(ctx, service, domain)
		gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())

		slices.Sort(instances)

		if slices.Equal(instances, names) {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}
}
