package service

import (
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
