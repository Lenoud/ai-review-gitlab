package service

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"
)

type CodeReviewReportInput struct {
	LogID   uint
	LogType string
	Token   string
}

type OpenReportRepository interface {
	FindPushByID(ctx context.Context, id uint) (*PushReviewLog, error)
	FindMergeRequestByID(ctx context.Context, id uint) (*MergeRequestReviewLog, error)
}

type OpenReportService struct {
	logs OpenReportRepository
	now  func() time.Time
}

func NewOpenReportService(logs OpenReportRepository) *OpenReportService {
	return &OpenReportService{logs: logs, now: time.Now}
}

func (s *OpenReportService) CodeReviewReport(ctx context.Context, input CodeReviewReportInput) (string, error) {
	input.LogType = strings.ToLower(strings.TrimSpace(input.LogType))
	input.Token = strings.TrimSpace(input.Token)
	if input.LogID == 0 || input.LogType == "" || input.Token == "" {
		return "", ErrInvalidReviewLogInput
	}
	switch input.LogType {
	case "push":
		log, err := s.logs.FindPushByID(ctx, input.LogID)
		if err != nil {
			return "", err
		}
		if !validShareToken(input.Token, log.ShareToken, log.ShareTokenExpiresAt, s.now()) {
			return "", ErrReviewLogNotFound
		}
		return buildPushReviewReportHTML(log), nil
	case "mr", "merge_request", "merge-request":
		log, err := s.logs.FindMergeRequestByID(ctx, input.LogID)
		if err != nil {
			return "", err
		}
		if !validShareToken(input.Token, log.ShareToken, log.ShareTokenExpiresAt, s.now()) {
			return "", ErrReviewLogNotFound
		}
		return buildMergeRequestReviewReportHTML(log), nil
	default:
		return "", ErrInvalidReviewLogInput
	}
}

func validShareToken(inputToken string, storedToken string, expiresAt int64, now time.Time) bool {
	return storedToken != "" && inputToken == storedToken && expiresAt > now.UnixMilli()
}

func buildPushReviewReportHTML(log *PushReviewLog) string {
	subtitle := fmt.Sprintf("Push · %s · %s", log.Branch, log.AuthorDisplayText)
	return buildCodeReviewReportHTML(codeReviewReportView{
		Title:        log.ProjectName,
		Subtitle:     subtitle,
		Score:        log.Score,
		Additions:    log.Additions,
		Deletions:    log.Deletions,
		ReviewResult: log.ReviewResult,
	})
}

func buildMergeRequestReviewReportHTML(log *MergeRequestReviewLog) string {
	subtitle := fmt.Sprintf("Merge Request · %s -> %s · %s", log.SourceBranch, log.TargetBranch, log.AuthorDisplayText)
	return buildCodeReviewReportHTML(codeReviewReportView{
		Title:        log.ProjectName,
		Subtitle:     subtitle,
		Score:        log.Score,
		Additions:    log.Additions,
		Deletions:    log.Deletions,
		ReviewResult: log.ReviewResult,
	})
}

type codeReviewReportView struct {
	Title        string
	Subtitle     string
	Score        int
	Additions    int
	Deletions    int
	ReviewResult string
}

func buildCodeReviewReportHTML(view codeReviewReportView) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s - AI Review</title>
  <style>
    body{margin:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#f5f7fb;color:#172033}
    main{max-width:960px;margin:0 auto;padding:40px 20px}
    header{border-bottom:1px solid #d9e0ee;padding-bottom:20px;margin-bottom:24px}
    h1{font-size:28px;line-height:1.2;margin:0 0 8px}
    .meta{color:#667085;font-size:14px}
    .stats{display:flex;gap:12px;flex-wrap:wrap;margin:24px 0}
    .stat{background:#fff;border:1px solid #d9e0ee;border-radius:8px;padding:14px 18px;min-width:110px}
    .label{font-size:12px;color:#667085;margin-bottom:6px}
    .value{font-size:24px;font-weight:700}
    article{background:#fff;border:1px solid #d9e0ee;border-radius:8px;padding:20px;white-space:pre-wrap;line-height:1.65}
  </style>
</head>
<body>
  <main>
    <header>
      <h1>%s</h1>
      <div class="meta">%s</div>
    </header>
    <section class="stats">
      <div class="stat"><div class="label">评分</div><div class="value">%d</div></div>
      <div class="stat"><div class="label">新增行</div><div class="value">+%d</div></div>
      <div class="stat"><div class="label">删除行</div><div class="value">-%d</div></div>
    </section>
    <article>%s</article>
  </main>
</body>
</html>`,
		html.EscapeString(view.Title),
		html.EscapeString(view.Title),
		html.EscapeString(view.Subtitle),
		view.Score,
		view.Additions,
		view.Deletions,
		html.EscapeString(strings.TrimSpace(view.ReviewResult)),
	)
}
