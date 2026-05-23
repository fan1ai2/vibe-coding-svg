package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/gin-gonic/gin"
)

func testHandler() *ConversionHandler {
	return &ConversionHandler{
		cfg: &config.Config{MaxFileSize: 10 << 20},
	}
}

func TestUploadNoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/conversions", nil)
	c.Set("user_id", "test-user")

	h := testHandler()
	h.Upload(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal("failed to decode response:", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error in response")
	}
	if errObj["code"] != "NO_FILE" {
		t.Errorf("expected code NO_FILE, got %v", errObj["code"])
	}
}

func TestUploadNoExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "test-user")

	// 构造不带扩展名的文件上传
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, _ := writer.CreateFormFile("file", "noextension")
	part.Write([]byte("fake"))
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/conversions", buf)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	h := testHandler()
	h.Upload(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for no extension, got %d", w.Code)
	}
}

func TestUploadInvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "test-user")

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write(bytes.Repeat([]byte("not an image at all"), 100))
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/conversions", buf)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	h := testHandler()
	h.Upload(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid type, got %d", w.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "INVALID_TYPE" {
		t.Errorf("expected code INVALID_TYPE, got %v", errObj["code"])
	}
}
