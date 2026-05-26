package service

import "strings"

const reviewScoreFormatInstruction = "\n\n### 输出要求（必须遵守）\n请在回答末尾给出总分，格式严格为：总分:XX分（例如：总分:85分），以便系统解析。\n"

const reviewContentTruncatedMarker = "\n\n[内容已按 token 上限截断]"

type ReviewPromptAssembleInput struct {
	Template       string
	Diffs          string
	Commits        string
	ProjectName    string
	MaxInputTokens int
}

func AssembleReviewPrompt(input ReviewPromptAssembleInput) string {
	replacer := strings.NewReplacer(
		"{{projectName}}", input.ProjectName,
		"{{diffs}}", "",
		"{{commits}}", "",
	)
	fixedPart := replacer.Replace(input.Template) + reviewScoreFormatInstruction
	contentPrefix := "\n\n### 待审查内容\n**代码变更内容**：\n"
	commitPrefix := "\n\n**提交历史（commits）**：\n"
	diffs := input.Diffs
	commits := input.Commits
	prompt := fixedPart + contentPrefix + diffs + commitPrefix + commits
	if input.MaxInputTokens <= 0 || CountReviewTokens(prompt) <= input.MaxInputTokens {
		return prompt
	}

	reserved := fixedPart + contentPrefix + commitPrefix
	budget := input.MaxInputTokens - CountReviewTokens(reserved) - CountReviewTokens(reviewContentTruncatedMarker)
	if budget <= 0 {
		return fixedPart + contentPrefix + strings.TrimSpace(reviewContentTruncatedMarker) + commitPrefix
	}
	commitBudget := budget / 5
	if commitBudget > 0 && CountReviewTokens(commits) > commitBudget {
		commits = TruncateReviewTextByTokens(commits, commitBudget) + reviewContentTruncatedMarker
	}
	diffBudget := input.MaxInputTokens - CountReviewTokens(fixedPart+contentPrefix+commitPrefix+commits) - CountReviewTokens(reviewContentTruncatedMarker)
	if diffBudget <= 0 {
		diffBudget = budget
	}
	if CountReviewTokens(diffs) > diffBudget {
		diffs = TruncateReviewTextByTokens(diffs, diffBudget) + reviewContentTruncatedMarker
	}
	return fixedPart + contentPrefix + diffs + commitPrefix + commits
}
