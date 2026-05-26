package service

import "strings"

const reviewScoreFormatInstruction = "\n\n### 输出要求（必须遵守）\n请在回答末尾给出总分，格式严格为：总分:XX分（例如：总分:85分），以便系统解析。\n"

type ReviewPromptAssembleInput struct {
	Template    string
	Diffs       string
	Commits     string
	ProjectName string
}

func AssembleReviewPrompt(input ReviewPromptAssembleInput) string {
	replacer := strings.NewReplacer(
		"{{projectName}}", input.ProjectName,
		"{{diffs}}", "",
		"{{commits}}", "",
	)
	fixedPart := replacer.Replace(input.Template) + reviewScoreFormatInstruction
	return fixedPart + "\n\n### 待审查内容\n**代码变更内容**：\n" + input.Diffs + "\n\n**提交历史（commits）**：\n" + input.Commits
}
