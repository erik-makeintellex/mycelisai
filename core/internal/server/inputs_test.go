package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/internal/inputs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleInputSourcesCreateAndBuffer(t *testing.T) {
	srv := &AdminServer{Inputs: inputs.NewService()}
	body := []byte(`{
		"id":"warehouse-sensor",
		"name":"Warehouse Sensor",
		"source_type":"sensor",
		"adapter_kind":"sensor",
		"buffer_mode":"latest_state",
		"scope_kind":"group",
		"scope_ref":"ops"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/input-sources", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.HandleInputSources(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var created protocol.APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !created.OK {
		t.Fatalf("created = %+v", created)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/input-sources/warehouse-sensor/buffer?mode=latest_state", nil)
	req.SetPathValue("id", "warehouse-sensor")
	rec = httptest.NewRecorder()
	srv.HandleInputSourceBuffer(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("buffer status = %d body = %s", rec.Code, rec.Body.String())
	}
	var view struct {
		OK   bool              `json:"ok"`
		Data inputs.BufferView `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &view); err != nil {
		t.Fatalf("decode buffer: %v", err)
	}
	if !view.OK || view.Data.Source.ID != "warehouse-sensor" || view.Data.Mode != inputs.BufferLatestState {
		t.Fatalf("view = %+v", view)
	}
}

func TestHandleInputSourcesRejectsRawSecret(t *testing.T) {
	srv := &AdminServer{Inputs: inputs.NewService()}
	body := []byte(`{
		"id":"vendor-api",
		"name":"Vendor API",
		"source_type":"api",
		"adapter_kind":"api",
		"auth_scheme":"api_key",
		"secret_ref":"plain-secret-value"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/input-sources", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.HandleInputSources(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

func TestHandleInputSourcesRejectsDuplicateIngressSubject(t *testing.T) {
	srv := &AdminServer{Inputs: inputs.NewService()}
	first := []byte(`{
		"id":"service-a",
		"name":"Service A",
		"allowed_ingress_subject":"swarm.global.input.shared-events"
	}`)
	second := []byte(`{
		"id":"service-b",
		"name":"Service B",
		"allowed_ingress_subject":"swarm.global.input.shared-events"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/input-sources", bytes.NewReader(first))
	rec := httptest.NewRecorder()
	srv.HandleInputSources(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first status = %d body = %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/input-sources", bytes.NewReader(second))
	rec = httptest.NewRecorder()
	srv.HandleInputSources(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d body = %s", rec.Code, rec.Body.String())
	}
}
