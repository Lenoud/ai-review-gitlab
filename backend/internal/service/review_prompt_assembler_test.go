package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReviewPromptAssemblerAssemblesTemplateDiffsAndCommits(t *testing.T) {
	prompt := AssembleReviewPrompt(ReviewPromptAssembleInput{
		Template:    "项目 {{projectName}} 请重点关注安全风险。{{diffs}}{{commits}}",
		Diffs:       "diff --git a/main.go b/main.go\n+fmt.Println(\"x\")",
		Commits:     "fix auth",
		ProjectName: "AI Review",
	})

	require.Contains(t, prompt, "项目 AI Review 请重点关注安全风险。")
	require.NotContains(t, prompt, "{{diffs}}")
	require.NotContains(t, prompt, "{{commits}}")
	require.Contains(t, prompt, "### 输出要求（必须遵守）")
	require.Contains(t, prompt, "总分:XX分")
	require.Contains(t, prompt, "### 待审查内容")
	require.Contains(t, prompt, "**代码变更内容**：\n")
	require.Contains(t, prompt, "diff --git a/main.go b/main.go")
	require.Contains(t, prompt, "**提交历史（commits）**：\nfix auth")
}

func TestReviewPromptAssemblerTruncatesReviewContentByTokenLimit(t *testing.T) {
	prompt := AssembleReviewPrompt(ReviewPromptAssembleInput{
		Template:       "项目 {{projectName}} 请重点关注安全风险。",
		Diffs:          strings.Repeat("diff-token ", 300),
		Commits:        strings.Repeat("commit-token ", 100),
		ProjectName:    "AI Review",
		MaxInputTokens: 180,
	})

	require.LessOrEqual(t, CountReviewTokens(prompt), 180)
	require.Contains(t, prompt, "项目 AI Review 请重点关注安全风险。")
	require.Contains(t, prompt, "### 输出要求（必须遵守）")
	require.Contains(t, prompt, "总分:XX分")
	require.Contains(t, prompt, "### 待审查内容")
	require.Contains(t, prompt, "**代码变更内容**：\n")
	require.NotContains(t, prompt, strings.Repeat("diff-token ", 300))
	require.Contains(t, prompt, "[内容已按 token 上限截断]")
}

func TestReviewPromptAssemblerDoesNotTruncateWhenLimitIsNotSet(t *testing.T) {
	diff := strings.Repeat("diff-token ", 120)
	prompt := AssembleReviewPrompt(ReviewPromptAssembleInput{
		Template:    "项目 {{projectName}} 请检查。",
		Diffs:       diff,
		ProjectName: "AI Review",
	})

	require.Contains(t, prompt, diff)
	require.NotContains(t, prompt, "[内容已按 token 上限截断]")
}

func TestReviewPromptAssemblerDropsReviewContentWhenLimitCannotFitIt(t *testing.T) {
	prompt := AssembleReviewPrompt(ReviewPromptAssembleInput{
		Template:       "项目 {{projectName}} 请检查。",
		Diffs:          strings.Repeat("diff-token ", 300),
		Commits:        strings.Repeat("commit-token ", 100),
		ProjectName:    "AI Review",
		MaxInputTokens: 40,
	})

	require.Contains(t, prompt, "项目 AI Review 请检查。")
	require.Contains(t, prompt, "总分:XX分")
	require.Contains(t, prompt, "[内容已按 token 上限截断]")
	require.NotContains(t, prompt, "diff-token")
	require.NotContains(t, prompt, "commit-token")
}
