package ai

import (
	"testing"
)

func TestParseSSE_ValidChunks(t *testing.T) {
	// Simulate multi-chunk SSE stream where content is assembled incrementally
	body := `data: {"choices":[{"delta":{"content":"["}}]}

data: {"choices":[{"delta":{"content":"hello"}}]}

data: [DONE]`

	result, err := parseSSE(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "[hello" {
		t.Errorf("expected '[hello', got %q", result)
	}
}

func TestParseSSE_NoDoneMarker(t *testing.T) {
	body := `data: {"choices":[{"delta":{"content":"hello"}}]}`
	result, err := parseSSE(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestParseSSE_EmptyInput(t *testing.T) {
	result, err := parseSSE("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestParseSSE_NonSSELines(t *testing.T) {
	body := "just some text\nnot data: format\n"
	result, err := parseSSE(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestParseSSE_MessageContentField(t *testing.T) {
	body := `data: {"choices":[{"message":{"content":"hello from message"}}]}`
	result, err := parseSSE(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello from message" {
		t.Errorf("expected 'hello from message', got %q", result)
	}
}

func TestParseSSE_TextField(t *testing.T) {
	body := `data: {"choices":[{"text":"hello from text"}]}`
	result, err := parseSSE(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello from text" {
		t.Errorf("expected 'hello from text', got %q", result)
	}
}

func TestParseNonStreaming_ContentField(t *testing.T) {
	body := `{"choices":[{"message":{"content":"hello world"}}]}`
	result := parseNonStreaming(body)
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestParseNonStreaming_TextField(t *testing.T) {
	body := `{"choices":[{"text":"hello text"}]}`
	result := parseNonStreaming(body)
	if result != "hello text" {
		t.Errorf("expected 'hello text', got %q", result)
	}
}

func TestParseNonStreaming_EmptyBody(t *testing.T) {
	result := parseNonStreaming("")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestParseNonStreaming_InvalidJSON(t *testing.T) {
	result := parseNonStreaming("{not valid")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestStripMarkdownFences_JsonFence(t *testing.T) {
	input := "```json\n[{\"name\":\"icon\"}]\n```"
	expected := `[{"name":"icon"}]`
	result := stripMarkdownFences(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStripMarkdownFences_PlainFence(t *testing.T) {
	input := "```\nsome text\n```"
	result := stripMarkdownFences(input)
	if result != "some text" {
		t.Errorf("expected 'some text', got %q", result)
	}
}

func TestStripMarkdownFences_NoFence(t *testing.T) {
	input := "plain text"
	result := stripMarkdownFences(input)
	if result != "plain text" {
		t.Errorf("expected 'plain text', got %q", result)
	}
}

func TestStripMarkdownFences_OnlyOpenFence(t *testing.T) {
	// If only opening fence, function should strip it but nothing else changes
	input := "```json\ncontent without close"
	result := stripMarkdownFences(input)
	// The function strips ```json prefix and trims space, but won't find closing ```
	if result != "content without close" {
		t.Errorf("got %q", result)
	}
}

func TestFilterCandidates_AllValid(t *testing.T) {
	candidates := []IconCandidate{
		{Name: "a", SvgContent: `<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M1 1h22v22H1z"/></svg>`, Tags: []string{"ui"}},
		{Name: "b", SvgContent: `<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`, Tags: []string{"shape"}},
	}
	result := FilterCandidates(candidates)
	if len(result) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(result))
	}
}

func TestFilterCandidates_TooLarge(t *testing.T) {
	large := `<svg viewBox="0 0 24 24">`
	for i := 0; i < 50*1024; i++ {
		large += " "
	}
	large += `</svg>`
	candidates := []IconCandidate{
		{Name: "big", SvgContent: large, Tags: nil},
	}
	result := FilterCandidates(candidates)
	if len(result) != 0 {
		t.Errorf("expected 0 (too large), got %d", len(result))
	}
}

func TestFilterCandidates_NoViewBox(t *testing.T) {
	candidates := []IconCandidate{
		{Name: "novb", SvgContent: `<svg xmlns="http://www.w3.org/2000/svg"><path d="M1 1h22v22H1z"/></svg>`, Tags: nil},
	}
	result := FilterCandidates(candidates)
	if len(result) != 0 {
		t.Errorf("expected 0 (no viewBox), got %d", len(result))
	}
}

func TestFilterCandidates_NotSVG(t *testing.T) {
	candidates := []IconCandidate{
		{Name: "nosvg", SvgContent: `<div>not svg</div>`, Tags: nil},
	}
	result := FilterCandidates(candidates)
	if len(result) != 0 {
		t.Errorf("expected 0 (not svg), got %d", len(result))
	}
}

func TestFilterCandidates_EmptyInput(t *testing.T) {
	result := FilterCandidates(nil)
	if len(result) != 0 {
		t.Errorf("expected 0 for nil, got %d", len(result))
	}
}

func TestDetectStyleMismatch_LineStyleNoMismatch(t *testing.T) {
	svg := `<svg viewBox="0 0 24 24"><path fill="none" stroke="currentColor" d="M1 1h22"/><circle fill="none" stroke="currentColor" cx="12" cy="12" r="10"/></svg>`
	if DetectStyleMismatch(svg, "line") {
		t.Error("expected no mismatch for line style with fill=none")
	}
}

func TestDetectStyleMismatch_LineStyleMismatch(t *testing.T) {
	svg := `<svg viewBox="0 0 24 24"><path fill="#000" d="M1 1h22"/><circle fill="#fff" cx="12" cy="12" r="10"/></svg>`
	if !DetectStyleMismatch(svg, "line") {
		t.Error("expected mismatch for line style with solid fills")
	}
}

func TestDetectStyleMismatch_FilledStyle(t *testing.T) {
	// Style "filled" should never trigger mismatch
	svg := `<svg viewBox="0 0 24 24"><path fill="#000" d="M1 1h22"/></svg>`
	if DetectStyleMismatch(svg, "filled") {
		t.Error("filled style should never trigger mismatch")
	}
}

func TestDetectStyleMismatch_NoFills(t *testing.T) {
	svg := `<svg viewBox="0 0 24 24"><g stroke="currentColor"><path d="M1 1h22"/></g></svg>`
	if DetectStyleMismatch(svg, "line") {
		t.Error("expected no mismatch when there are no fill attributes")
	}
}
