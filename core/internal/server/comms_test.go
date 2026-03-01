package server

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/comms"
)

type testCommsProvider struct {
	info     comms.ProviderInfo
	lastReq  comms.SendRequest
	sendErr  error
	sendResp comms.SendResult
}

func (p *testCommsProvider) Info() comms.ProviderInfo { return p.info }

func (p *testCommsProvider) Send(_ context.Context, req comms.SendRequest) (comms.SendResult, error) {
	p.lastReq = req
	if p.sendErr != nil {
		return comms.SendResult{}, p.sendErr
	}
	return p.sendResp, nil
}

func withComms(provider comms.Provider) func(*AdminServer) {
	return func(s *AdminServer) {
		g := comms.NewGateway()
		g.Register(provider)
		s.Comms = g
	}
}

func TestHandleCommsProviders_HappyPath(t *testing.T) {
	s := newTestServer(withComms(&testCommsProvider{
		info: comms.ProviderInfo{Name: "slack", Channel: "chat", Description: "Slack", Configured: true},
	}))

	rr := doRequest(t, http.HandlerFunc(s.HandleCommsProviders), "GET", "/api/v1/comms/providers", "")
	assertStatus(t, rr, http.StatusOK)
	if !strings.Contains(rr.Body.String(), "slack") {
		t.Fatalf("expected provider in response: %s", rr.Body.String())
	}
}

func TestHandleCommsSend_HappyPath(t *testing.T) {
	prov := &testCommsProvider{
		info: comms.ProviderInfo{Name: "slack", Channel: "chat", Description: "Slack", Configured: true},
		sendResp: comms.SendResult{
			Provider: "slack",
			Status:   "sent",
		},
	}
	s := newTestServer(withComms(prov))
	body := `{"provider":"slack","recipient":"#ops","message":"health check"}`
	rr := doRequest(t, http.HandlerFunc(s.HandleCommsSend), "POST", "/api/v1/comms/send", body)
	assertStatus(t, rr, http.StatusOK)
	if prov.lastReq.Message != "health check" {
		t.Fatalf("provider did not receive message")
	}
}

func TestHandleCommsSend_ValidationAndOffline(t *testing.T) {
	s := newTestServer()
	rr := doRequest(t, http.HandlerFunc(s.HandleCommsSend), "POST", "/api/v1/comms/send", `{"provider":"slack","message":"x"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)

	prov := &testCommsProvider{
		info:    comms.ProviderInfo{Name: "slack", Channel: "chat", Description: "Slack", Configured: true},
		sendErr: context.DeadlineExceeded,
	}
	s = newTestServer(withComms(prov))
	rr = doRequest(t, http.HandlerFunc(s.HandleCommsSend), "POST", "/api/v1/comms/send", `{"provider":"slack","message":"x"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleCommsInbound_NATSOffline(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/comms/inbound/{provider}", s.HandleCommsInbound)
	rr := doRequest(t, mux, "POST", "/api/v1/comms/inbound/whatsapp", `{"sender":"+1","message":"hello"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}
