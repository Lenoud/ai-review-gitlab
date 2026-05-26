package service

import (
	"regexp"
	"strings"
)

var (
	thinkBlockPattern      = regexp.MustCompile(`(?is)<think\b[^>]*>.*?</think>`)
	unclosedThinkPattern   = regexp.MustCompile(`(?is)<think\b[^>]*>.*$`)
	reasoningHeaderPattern = regexp.MustCompile(`(?m)^\s*(思考过程|推理过程|思考|推理)\s*[:：].*(\n\s*)?`)
)

func CleanReviewOutput(output string) string {
	output = thinkBlockPattern.ReplaceAllString(output, "")
	output = unclosedThinkPattern.ReplaceAllString(output, "")
	output = reasoningHeaderPattern.ReplaceAllString(output, "")
	return strings.TrimSpace(output)
}
