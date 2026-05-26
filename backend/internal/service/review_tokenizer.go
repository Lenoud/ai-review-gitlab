package service

import (
	"sync"
	"unicode"
	"unicode/utf8"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

const reviewTokenEncodingName = "cl100k_base"

var reviewTokenizer = struct {
	sync.Once
	encoding *tiktoken.Tiktoken
	err      error
}{}

func CountReviewTokens(text string) int {
	encoding, err := getReviewTokenEncoding()
	if err == nil {
		return len(encoding.Encode(text, nil, nil))
	}
	return estimateReviewTokens(text)
}

func TruncateReviewTextByTokens(text string, maxTokens int) string {
	if maxTokens <= 0 || text == "" {
		return ""
	}
	encoding, err := getReviewTokenEncoding()
	if err == nil {
		tokens := encoding.Encode(text, nil, nil)
		if len(tokens) <= maxTokens {
			return text
		}
		return encoding.Decode(tokens[:maxTokens])
	}
	return truncateReviewTextByEstimatedTokens(text, maxTokens)
}

func getReviewTokenEncoding() (*tiktoken.Tiktoken, error) {
	reviewTokenizer.Do(func() {
		reviewTokenizer.encoding, reviewTokenizer.err = tiktoken.GetEncoding(reviewTokenEncodingName)
	})
	return reviewTokenizer.encoding, reviewTokenizer.err
}

func estimateReviewTokens(text string) int {
	if text == "" {
		return 0
	}
	chineseChars := 0
	otherChars := 0
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			chineseChars++
			continue
		}
		otherChars++
	}
	tokens := chineseChars/2 + otherChars/4
	if tokens < 1 {
		return 1
	}
	return tokens
}

func truncateReviewTextByEstimatedTokens(text string, maxTokens int) string {
	if maxTokens <= 0 || text == "" {
		return ""
	}
	chineseChars := 0
	otherChars := 0
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			chineseChars++
			continue
		}
		otherChars++
	}
	charsPerToken := 4
	if chineseChars > otherChars {
		charsPerToken = 2
	}
	maxChars := maxTokens * charsPerToken
	if utf8.RuneCountInString(text) <= maxChars {
		return text
	}
	out := make([]rune, 0, maxChars)
	for _, r := range text {
		if len(out) >= maxChars {
			break
		}
		out = append(out, r)
	}
	return string(out)
}
