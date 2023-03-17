package providertest

import (
	"context"
	"reflect"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/onsi/gomega"
	"golang.org/x/exp/slices"
)

func expectInstanceListToEventuallyEqual(
	ctx context.Context,
	res *dnssd.UnicastResolver,
	service, domain string,
	expect ...dnssd.ServiceInstance,
) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	var names []string
	for _, inst := range expect {
		names = append(names, inst.Instance)
	}

	slices.Sort(names)

	var previous []string

	for {
		instances, err := res.EnumerateInstances(ctx, service, domain)
		if ctx.Err() != nil {
			gomega.ExpectWithOffset(1, previous).To(
				gomega.ConsistOf(names),
				"timed-out waiting for convergence",
			)
		}
		gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())

		slices.Sort(instances)

		if slices.Equal(instances, names) {
			return
		}

		previous = instances
		time.Sleep(10 * time.Millisecond)
	}
}

func expectInstanceToEventuallyEqual(
	ctx context.Context,
	res *dnssd.UnicastResolver,
	expect dnssd.ServiceInstance,
) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	var previous dnssd.ServiceInstance

	for {
		actual, ok, err := res.LookupInstance(
			ctx,
			expect.Instance,
			expect.ServiceType,
			expect.Domain,
		)
		if ctx.Err() != nil {
			gomega.ExpectWithOffset(1, previous).To(
				gomega.Equal(expect),
				"timed-out waiting for convergence",
			)
		}
		gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())

		if ok && reflect.DeepEqual(actual, expect) {
			return
		}

		previous = actual
		time.Sleep(10 * time.Millisecond)
	}
}

func expectInstanceToEventuallyNotExist(
	ctx context.Context,
	res *dnssd.UnicastResolver,
	expect dnssd.ServiceInstance,
) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

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

		time.Sleep(10 * time.Millisecond)
	}
}
