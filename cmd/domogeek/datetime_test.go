package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDateTimeHandler_ServeHTTP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/datetime", nil)
	rec := httptest.NewRecorder()

	(&DateTimeHandler{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got CurrentDateTime
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}

	now := time.Now().In(location)
	if diff := now.Sub(got.DateTime); diff < -time.Second || diff > time.Second {
		t.Errorf("DateTime = %v, want close to %v", got.DateTime, now)
	}
	_, wantOffset := now.Zone()
	_, gotOffset := got.DateTime.Zone()
	if gotOffset != wantOffset {
		t.Errorf("DateTime UTC offset = %d, want %d", gotOffset, wantOffset)
	}
}
