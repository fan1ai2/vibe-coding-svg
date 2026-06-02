package service

import (
	"context"
	"fmt"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/ai"
)

type AiService struct {
	provider ai.Provider
	quota    *ai.QuotaService
}

func NewAiService(provider ai.Provider, quota *ai.QuotaService) *AiService {
	return &AiService{provider: provider, quota: quota}
}

type GenerateRequest struct {
	Prompt string `json:"prompt" binding:"required,min=1,max=200"`
	Style  string `json:"style"`
}

type GenerateResponse struct {
	Candidates     []ai.IconCandidate `json:"candidates"`
	RemainingQuota int                `json:"remaining_quota"`
}

type QuotaResponse struct {
	Remaining int `json:"remaining"`
	Limit     int `json:"limit"`
}

func (s *AiService) Generate(userID string, req GenerateRequest) (*GenerateResponse, error) {
	remaining, ok := s.quota.Check(userID)
	if !ok {
		return nil, fmt.Errorf("quota check failed")
	}
	if remaining <= 0 {
		return nil, fmt.Errorf("daily quota exceeded")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	candidates, err := s.provider.Generate(ctx, req.Prompt, req.Style)
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	if err := s.quota.Consume(userID); err != nil {
		return nil, fmt.Errorf("quota consume: %w", err)
	}

	remaining, _ = s.quota.Check(userID)

	return &GenerateResponse{
		Candidates:     candidates,
		RemainingQuota: remaining,
	}, nil
}

func (s *AiService) Quota(userID string) QuotaResponse {
	remaining, ok := s.quota.Check(userID)
	if !ok {
		remaining = 0
	}
	return QuotaResponse{Remaining: remaining, Limit: 20}
}
