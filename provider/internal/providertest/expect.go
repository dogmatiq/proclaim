package providertest

import (
	"context"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/onsi/ginkgo/v2"
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
		names = append(names, inst.Name)
	}

	slices.Sort(names)

	var previous []string

	for {
		instances, err := res.EnumerateInstances(ctx, service, domain)
		switch err {
		case context.DeadlineExceeded:
			if err == ctx.Err() {
				gomega.ExpectWithOffset(1, previous).To(
					gomega.ConsistOf(names),
					"timed-out waiting for instance list to converge",
				)
			}
		default:
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
		case nil:
			slices.Sort(instances)
			if slices.Equal(instances, names) {
				return
			}
			previous = instances
		}

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
			expect.Name,
			expect.ServiceType,
			expect.Domain,
		)
		switch err {
		case context.DeadlineExceeded:
			if err == ctx.Err() {
				gomega.ExpectWithOffset(1, previous).To(
					gomega.Equal(expect),
					"timed-out waiting for instance to converge on positive result",
				)
			}
		default:
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
		case nil:
			if ok && actual.Equal(expect) {
				return
			}
			previous = actual
		}

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
			expect.Name,
			expect.ServiceType,
			expect.Domain,
		)
		switch err {
		case context.DeadlineExceeded:
			if ctx.Err() == err {
				ginkgo.Fail("timed-out wiating for instance to converge on negative result")
			}
		default:
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
		case nil:
			if !ok {
				return
			}
		}

		time.Sleep(10 * time.Millisecond)
	}
}
