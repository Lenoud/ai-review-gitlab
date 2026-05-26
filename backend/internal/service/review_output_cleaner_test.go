package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanReviewOutputRemovesThinkTags(t *testing.T) {
	output := CleanReviewOutput("<think>\n先分析一下风险\n</think>\n\n#### 一、关键问题\n- 缺少鉴权\n\n总分:72分")

	require.Equal(t, "#### 一、关键问题\n- 缺少鉴权\n\n总分:72分", output)
}

func TestCleanReviewOutputRemovesUnclosedThinkTag(t *testing.T) {
	output := CleanReviewOutput("前置说明\n<think>这里是未闭合的推理内容")

	require.Equal(t, "前置说明", output)
}

func TestCleanReviewOutputRemovesReasoningSections(t *testing.T) {
	output := CleanReviewOutput("思考过程：先看 diff，再判断风险。\n\n推理过程：这里不应该展示。\n\n#### 一、关键问题\n- SQL 注入\n\n总分:65分")

	require.Equal(t, "#### 一、关键问题\n- SQL 注入\n\n总分:65分", output)
}

func TestCleanReviewOutputKeepsReviewWhenReasoningLineHasNoBlankSeparator(t *testing.T) {
	output := CleanReviewOutput("思考过程：简短判断\n#### 一、关键问题\n- SQL 注入\n\n总分:65分")

	require.Equal(t, "#### 一、关键问题\n- SQL 注入\n\n总分:65分", output)
}
