package forwarder

import (
	"context"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/kloud-team/dns/config"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

type dnsHandler struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

	h.logger.Debug("serveDNS: start")
	defer h.logger.Debug("serveDNS: end")

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = false

	var clientIP = ""
	if len(r.Question) > 0 {

		q := r.Question[0]
		clientIP = getClientIP(w)
		h.logger.Debugf("serveDNS: clientIP=%s, q.Name=%s, q.Qtype=%d", clientIP, q.Name, q.Qtype)

		newResolver := func(addr string) *net.Resolver {
			return &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: time.Second * 5,
					}
					host := fmt.Sprintf("%s:%d", addr, 53)
					h.logger.Infof("resolve with addr=%s", host)
					return d.DialContext(ctx, "udp", host)
				},
			}
		}

		for _, forward := range h.cfg.Forwards {

			if forward.Templates != nil {
				if len(forward.Templates.Rewrites) > 0 {
					for _, rwr := range forward.Templates.Rewrites {
						re := regexp.MustCompile(rwr.Name)
						if re.MatchString(q.Name) {
							q.Name = fmt.Sprintf("%scluster.local", q.Name)
						}
					}
				}
			}

			res := newResolver(forward.Name)
			answers, err := h.answer(res, q, r)
			if err != nil {
				h.logger.Debugf("could not get answers: %+v", err)
				continue
			}
			msg.Answer = append(msg.Answer, answers...)
		}

	}

	// Handle response writing separately from query metrics
	if err := w.WriteMsg(msg); err != nil {
		h.logger.Debugf("failed to write DNS response: %v", err)
	}
}

func (h *dnsHandler) answer(res *net.Resolver, q dns.Question, r *dns.Msg) ([]dns.RR, error) {

	cname, err := res.LookupCNAME(context.Background(), q.Name)
	if err != nil {
		return nil, fmt.Errorf("could not lookupCNAME: %+v", err)
	}

	if strings.Contains(cname, "cluster.local") {
		cname = strings.ReplaceAll(cname, ".cluster.local", "")
	}

	log.Println(cname)

	//r.Answer = append(r.Answer, &dns.CNAME{
	//Hdr: dns.RR_Header{
	//Name:   q.Name,
	//Rrtype: dns.TypeCNAME,
	//Class:  dns.ClassINET,
	//},
	//Target: cname,
	//})

	ips, err := res.LookupHost(context.Background(), q.Name)
	if err != nil {
		return nil, fmt.Errorf("could not lookupHOST: %+v", err)
	}

	for _, ip := range ips {

		if !isIPv4(ip) {
			continue
		}

		r.Answer = append(r.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   cname,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			A: net.ParseIP(ip),
		})
	}

	h.logger.Debugf("serveDNS: got answers for %s", q.Name)

	return r.Answer, nil
}

func Start(ctx context.Context, cfg *config.Config, logger *zap.SugaredLogger) error {

	handler := &dnsHandler{logger: logger, cfg: cfg}
	server := &dns.Server{
		Addr:      fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Net:       "udp",
		Handler:   handler,
		UDPSize:   65535,
		ReusePort: true,
	}

	logger.Infof("starting DNS server on port %d (UDP)", cfg.Port)

	return server.ListenAndServe()
}

func getClientIP(w dns.ResponseWriter) string {
	addr := w.RemoteAddr()
	switch a := addr.(type) {
	case *net.UDPAddr:
		return a.IP.String()
	case *net.TCPAddr:
		return a.IP.String()
	default:
		s := addr.String()
		if i := strings.LastIndex(s, ":"); i > 0 {
			return s[:i]
		}
		return s
	}
}

func isIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}
