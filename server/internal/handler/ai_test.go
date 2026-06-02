package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

type mockAiGenerator struct {
	generateResp *service.GenerateResponse
	generateErr  error
}

func (m *mockAiGenerator) Generate(userID string, req service.GenerateRequest) (*service.GenerateResponse, error) {
	return m.generateResp, m.generateErr
}

func (m *mockAiGenerator) Quota(userID string) service.QuotaResponse {
	return service.QuotaResponse{Remaining: 10, Limit: 20}
}

func newTestHandler(mock *mockAiGenerator) *AiHandler {
	return &AiHandler{aiService: mock}
}

func ginCtx(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "test-user")
	return c, w
}

func TestAiGenerate_MissingPrompt(t *testing.T) {
	mock := &mockAiGenerator{
		generateResp: nil,
		generateErr:  nil,
	}
	h := newTestHandler(mock)
	c, w := ginCtx("POST", "/ai/generate", toJSON(service.GenerateRequest{Prompt: ""}))

	h.Generate(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "INVALID_PARAMS" {
		t.Errorf("expected INVALID_PARAMS, got %v", errObj["code"])
	}
}

func TestAiGenerate_PromptTooLong(t *testing.T) {
	mock := &mockAiGenerator{}
	h := newTestHandler(mock)

	longPrompt := ""
	for i := 0; i < 201; i++ {
		longPrompt += "x"
	}
	c, w := ginCtx("POST", "/ai/generate", toJSON(service.GenerateRequest{Prompt: longPrompt}))

	h.Generate(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "INVALID_PARAMS" {
		t.Errorf("expected INVALID_PARAMS, got %v", errObj["code"])
	}
}

func TestAiGenerate_DefaultStyle(t *testing.T) {
	mock := &mockAiGenerator{
		generateResp: &service.GenerateResponse{
			Candidates:     nil,
			RemainingQuota: 19,
		},
	}
	h := newTestHandler(mock)
	c, w := ginCtx("POST", "/ai/generate", toJSON(service.GenerateRequest{Prompt: "a cat icon"}))

	h.Generate(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAiGenerate_InvalidStyle(t *testing.T) {
	mock := &mockAiGenerator{}
	h := newTestHandler(mock)
	// binding: "oneof=line filled" will reject "gradient"
	c, w := ginCtx("POST", "/ai/generate", toJSON(map[string]string{"prompt": "test", "style": "gradient"}))

	h.Generate(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid style, got %d", w.Code)
	}
}

func TestAiGenerate_QuotaExceeded(t *testing.T) {
	mock := &mockAiGenerator{
		generateErr: errMsg("daily quota exceeded"),
	}
	h := newTestHandler(mock)
	c, w := ginCtx("POST", "/ai/generate", toJSON(service.GenerateRequest{Prompt: "test"}))

	h.Generate(c)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "QUOTA_EXCEEDED" {
		t.Errorf("expected QUOTA_EXCEEDED, got %v", errObj["code"])
	}
}

func TestAiGenerate_ServiceError(t *testing.T) {
	mock := &mockAiGenerator{
		generateErr: errMsg("api timeout"),
	}
	h := newTestHandler(mock)
	c, w := ginCtx("POST", "/ai/generate", toJSON(service.GenerateRequest{Prompt: "test"}))

	h.Generate(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "GENERATE_FAILED" {
		t.Errorf("expected GENERATE_FAILED, got %v", errObj["code"])
	}
}

func TestAiQuota_Success(t *testing.T) {
	mock := &mockAiGenerator{}
	h := newTestHandler(mock)
	c, w := ginCtx("GET", "/ai/quota", nil)

	h.Quota(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	data := body["data"].(map[string]interface{})
	if data["remaining"].(float64) != 10 {
		t.Errorf("expected remaining 10, got %v", data["remaining"])
	}
}

func errMsg(msg string) error {
	return &stringError{msg}
}

type stringError struct{ msg string }

func (e *stringError) Error() string { return e.msg }

func toJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
