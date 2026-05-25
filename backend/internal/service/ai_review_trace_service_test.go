package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAIReviewTraceServiceCreateNormalizesInput(t *testing.T) {
	traces := &fakeAIReviewTraceRepository{}
	service := NewAIReviewTraceService(traces)

	trace, err := service.Create(context.Background(), AIReviewTraceInput{
		ReviewEventType: "MR",
		ReviewEventID:   202,
		Prompt:          "prompt",
		Response:        "response",
		Provider:        " openai ",
		ModelCode:       " gpt-test ",
	})

	require.NoError(t, err)
	require.Equal(t, uint(1), trace.ID)
	require.Equal(t, "merge_request", traces.lastCreate.ReviewEventType)
	require.Equal(t, "openai", traces.lastCreate.Provider)
	require.Equal(t, "gpt-test", traces.lastCreate.ModelCode)
}

func TestAIReviewTraceServiceRejectsInvalidInput(t *testing.T) {
	traces := &fakeAIReviewTraceRepository{}
	service := NewAIReviewTraceService(traces)

	_, err := service.Create(context.Background(), AIReviewTraceInput{
		ReviewEventType: "unknown",
		ReviewEventID:   1,
		Provider:        "openai",
		ModelCode:       "gpt-test",
	})

	require.ErrorIs(t, err, ErrInvalidAIReviewTraceInput)

	_, err = service.Get(context.Background(), "push", 0)
	require.ErrorIs(t, err, ErrInvalidAIReviewTraceInput)
}

type fakeAIReviewTraceRepository struct {
	lastCreate AIReviewTraceInput
}

func (r *fakeAIReviewTraceRepository) Create(ctx context.Context, input AIReviewTraceInput) (*AIReviewTrace, error) {
	r.lastCreate = input
	return &AIReviewTrace{ID: 1, ReviewEventType: input.ReviewEventType, ReviewEventID: input.ReviewEventID}, nil
}

func (r *fakeAIReviewTraceRepository) FindByReviewEvent(ctx context.Context, eventType string, eventID uint) (*AIReviewTrace, error) {
	return &AIReviewTrace{ID: 2, ReviewEventType: eventType, ReviewEventID: eventID}, nil
}
