package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wyw14/cry-094/internal/adapter/local"
	"github.com/wyw14/cry-094/internal/application/analysisservice"
	"github.com/wyw14/cry-094/internal/application/authservice"
	"github.com/wyw14/cry-094/internal/application/precheckservice"
	"github.com/wyw14/cry-094/internal/application/teamservice"
	"github.com/wyw14/cry-094/internal/application/upload"
	"github.com/wyw14/cry-094/internal/domain/identity"
	appmiddleware "github.com/wyw14/cry-094/internal/middleware"
	"github.com/wyw14/cry-094/internal/parser"
	"github.com/wyw14/cry-094/internal/platform/clock"
	"github.com/wyw14/cry-094/internal/platform/files"
	"github.com/wyw14/cry-094/internal/repository/memory"
)

type httpIDs struct{ value int }

func (i *httpIDs) NewID() string {
	i.value++
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", i.value)
}
func testRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	teams := memory.NewTeamRepo()
	artifacts := memory.NewArtifactRepo()
	analyses := memory.NewAnalysisRepo()
	plans := memory.NewPlanRepo()
	audits := memory.NewAuditRepo()
	objects := files.New()
	now := clock.Fixed{Value: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	ids := &httpIDs{}
	user, err := identity.Register("owner", "owner@example.test", "secret-pass", now.Now())
	if err != nil {
		t.Fatal(err)
	}
	identities := memory.NewIdentityRepo(user)
	secret := []byte("0123456789abcdef0123456789abcdef")
	signer, err := local.NewSigner("test-key", []byte("0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	authApp := authservice.New(identities, now, ids, secret, 15*time.Minute, time.Hour)
	router := NewRouter(NewHandler(teamservice.New(teams, audits, now, ids), upload.New(teams, artifacts, objects, now, ids), analysisservice.New(teams, artifacts, analyses, objects, parser.Factory{BuildID: "v1"}, now, ids), precheckservice.New(analyses, plans, teams, signer, now, ids), authApp), appmiddleware.NewAuth(secret).Require())
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "owner", "exp": time.Now().Add(time.Hour).Unix()}).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return router, token
}
func TestCreateTeamValidationIncludesRequestID(t *testing.T) {
	router, token := testRouter(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"x"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", "req-test")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errorBody := body["error"].(map[string]any)
	if errorBody["request_id"] != "req-test" || errorBody["code"] != "VALIDATION_ERROR" {
		t.Fatalf("unexpected error: %#v", errorBody)
	}
}
func TestCreateTeamAndSecurityHeaders(t *testing.T) {
	router, token := testRouter(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"Operations","owner_id":"owner"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
	if response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Header().Get("X-Request-ID") == "" {
		t.Fatal("security and request headers must be present")
	}
}
