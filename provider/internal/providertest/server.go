package providertest

import (
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/miekg/dns"
)

type server struct {
	done chan struct{}

	m       sync.Mutex
	records map[string]map[uint16][]dns.RR
}

func startServer() (*server, *dnssd.UnicastResolver, error) {
	s := &server{
		done: make(chan struct{}),
	}

	host := "127.0.0.1"
	port := strconv.Itoa(5300 + rand.Intn(100))

	go s.run(host, port)

	return s, &dnssd.UnicastResolver{
		Config: &dns.ClientConfig{
			Servers: []string{host},
			Port:    port,
		},
	}, nil
}

func (s *server) SetRecords(records []dns.RR) {
	s.m.Lock()
	defer s.m.Unlock()

	s.records = map[string]map[uint16][]dns.RR{}

	for _, rr := range records {
		h := rr.Header()

		domainRecords := s.records[h.Name]
		if domainRecords == nil {
			domainRecords = map[uint16][]dns.RR{}
			s.records[h.Name] = domainRecords
		}

		domainRecords[h.Rrtype] = append(domainRecords[h.Rrtype], rr)
	}
}

func (s *server) Stop() {
	close(s.done)
}

// Run runs the server until ctx is canceled or an error occurs.
func (s *server) run(host, port string) {
	timeout := 5 * time.Second

	server := &dns.Server{
		Net:          "udp4",
		Addr:         net.JoinHostPort(host, port),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		Handler: dns.HandlerFunc(
			func(w dns.ResponseWriter, req *dns.Msg) {
				defer w.Close()

				if res, ok := s.buildResponse(req); ok {
					if err := w.WriteMsg(res); err != nil {
						panic(err)
					}
				}
			},
		),
	}

	go func() {
		<-s.done
		server.Shutdown()
	}()

	err := server.ListenAndServe()

	select {
	case <-s.done:
	default:
		if err != nil {
			panic(err)
		}
	}
}

// buildResponse builds the response to send in reply to the given request.
func (s *server) buildResponse(req *dns.Msg) (*dns.Msg, bool) {
	// We only support queries with exactly one question. The RFC allows for
	// multiple, but in practice this is nonsensical.
	//
	// See https://stackoverflow.com/questions/4082081/requesting-a-and-aaaa-records-in-single-dns-query/4085631#4085631
	// See https://www.rfc-editor.org/rfc/rfc1035
	if len(req.Question) != 1 {
		return nil, false
	}

	q := req.Question[0]

	res := &dns.Msg{}
	res.SetReply(req)
	res.Authoritative = true
	res.RecursionAvailable = false

	if q.Qclass != dns.ClassINET && q.Qclass != dns.ClassANY {
		res.Rcode = dns.RcodeNameError
		return res, true
	}

	s.m.Lock()
	records := s.records[q.Name]
	s.m.Unlock()

	if len(records) == 0 {
		res.Rcode = dns.RcodeNameError
		return res, true
	}

	// Always use a copy of the records in res.Answer.
	//
	// We don't want to reference the original slice(s) from s.records as they
	// may be modified as soon as s.m is unlocked.
	if q.Qtype == dns.TypeANY {
		for _, recs := range records {
			res.Answer = append(res.Answer, recs...)
		}
	} else {
		res.Answer = append([]dns.RR{}, records[q.Qtype]...)
	}

	return res, true
}
