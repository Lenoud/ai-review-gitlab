package service

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrAIReviewTraceNotFound     = errors.New("ai review trace not found")
	ErrInvalidAIReviewTraceInput = errors.New("invalid ai review trace input")
)

type AIReviewTrace struct {
	ID              uint   `json:"id"`
	ReviewEventType string `json:"reviewEventType"`
	ReviewEventID   uint   `json:"reviewEventId"`
	Prompt          string `json:"prompt"`
	Response        string `json:"response"`
	Provider        string `json:"provider"`
	ModelCode       string `json:"modelCode"`
	CreatedAt       int64  `json:"createdAt"`
	UpdatedAt       int64  `json:"updatedAt"`
}

type AIReviewTraceInput struct {
	ReviewEventType string
	ReviewEventID   uint
	Prompt          string
	Response        string
	Provider        string
	ModelCode       string
}

type AIReviewTraceRepository interface {
	Create(ctx context.Context, input AIReviewTraceInput) (*AIReviewTrace, error)
	FindByReviewEvent(ctx context.Context, eventType string, eventID uint) (*AIReviewTrace, error)
}

type AIReviewTraceService struct {
	traces AIReviewTraceRepository
}

func NewAIReviewTraceService(traces AIReviewTraceRepository) *AIReviewTraceService {
	return &AIReviewTraceService{traces: traces}
}

func (s *AIReviewTraceService) Create(ctx context.Context, input AIReviewTraceInput) (*AIReviewTrace, error) {
	normalized, err := normalizeAIReviewTraceInput(input)
	if err != nil {
		return nil, err
	}
	return s.traces.Create(ctx, normalized)
}

func (s *AIReviewTraceService) Get(ctx context.Context, eventType string, eventID uint) (*AIReviewTrace, error) {
	normalizedEventType, err := NormalizeReviewEventType(eventType)
	if err != nil || eventID == 0 {
		return nil, ErrInvalidAIReviewTraceInput
	}
	return s.traces.FindByReviewEvent(ctx, normalizedEventType, eventID)
}

func normalizeAIReviewTraceInput(input AIReviewTraceInput) (AIReviewTraceInput, error) {
	eventType, err := NormalizeReviewEventType(input.ReviewEventType)
	if err != nil || input.ReviewEventID == 0 {
		return AIReviewTraceInput{}, ErrInvalidAIReviewTraceInput
	}
	input.Provider = strings.TrimSpace(input.Provider)
	input.ModelCode = strings.TrimSpace(input.ModelCode)
	if input.Provider == "" || input.ModelCode == "" {
		return AIReviewTraceInput{}, ErrInvalidAIReviewTraceInput
	}
	input.ReviewEventType = eventType
	return input, nil
}

func NormalizeReviewEventType(eventType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(eventType)) {
	case ReviewTaskEventPush:
		return ReviewTaskEventPush, nil
	case "mr", "merge-request", ReviewTaskEventMergeRequest:
		return ReviewTaskEventMergeRequest, nil
	default:
		return "", ErrInvalidAIReviewTraceInput
	}
}
