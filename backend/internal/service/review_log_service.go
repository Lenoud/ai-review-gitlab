package service

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrReviewLogNotFound     = errors.New("review log not found")
	ErrInvalidReviewLogInput = errors.New("invalid review log input")
)

type ReviewCommit struct {
	Author    string `json:"author"`
	Message   string `json:"message"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
}

type PushReviewLog struct {
	ID                  uint           `json:"id"`
	ProjectID           uint           `json:"projectId"`
	ProjectName         string         `json:"projectName"`
	Author              string         `json:"author"`
	AuthorIdentity      string         `json:"authorIdentity"`
	AuthorDisplayName   string         `json:"authorDisplayName"`
	AuthorDisplayText   string         `json:"authorDisplayText"`
	Branch              string         `json:"branch"`
	CommitMessages      string         `json:"commitMessages"`
	Commits             []ReviewCommit `json:"commits"`
	Score               int            `json:"score"`
	Additions           int            `json:"additions"`
	Deletions           int            `json:"deletions"`
	LastCommitURL       string         `json:"lastCommitUrl"`
	ReviewResult        string         `json:"reviewResult"`
	ShareToken          string         `json:"shareToken"`
	ShareTokenExpiresAt int64          `json:"shareTokenExpiresAt"`
	CreatedAt           int64          `json:"createdAt"`
	UpdatedAt           int64          `json:"updatedAt"`
}

type MergeRequestReviewLog struct {
	ID                  uint   `json:"id"`
	ProjectID           uint   `json:"projectId"`
	ProjectName         string `json:"projectName"`
	Author              string `json:"author"`
	AuthorIdentity      string `json:"authorIdentity"`
	AuthorDisplayName   string `json:"authorDisplayName"`
	AuthorDisplayText   string `json:"authorDisplayText"`
	CommitMessages      string `json:"commitMessages"`
	Score               int    `json:"score"`
	SourceBranch        string `json:"sourceBranch"`
	TargetBranch        string `json:"targetBranch"`
	Additions           int    `json:"additions"`
	Deletions           int    `json:"deletions"`
	LastCommitID        string `json:"lastCommitId"`
	URL                 string `json:"url"`
	ReviewResult        string `json:"reviewResult"`
	ShareToken          string `json:"shareToken"`
	ShareTokenExpiresAt int64  `json:"shareTokenExpiresAt"`
	CreatedAt           int64  `json:"createdAt"`
	UpdatedAt           int64  `json:"updatedAt"`
}

type PushReviewLogInput struct {
	ProjectID         uint
	ProjectName       string
	Author            string
	AuthorIdentity    string
	AuthorDisplayName string
	Branch            string
	CommitMessages    string
	Commits           []ReviewCommit
	Score             int
	Additions         int
	Deletions         int
	LastCommitURL     string
	ReviewResult      string
}

type MergeRequestReviewLogInput struct {
	ProjectID         uint
	ProjectName       string
	Author            string
	AuthorIdentity    string
	AuthorDisplayName string
	SourceBranch      string
	TargetBranch      string
	CommitMessages    string
	Score             int
	Additions         int
	Deletions         int
	LastCommitID      string
	URL               string
	ReviewResult      string
}

type ReviewLogSearchQuery struct {
	ProjectID      uint
	Authors        []string
	ProjectNames   []string
	Branch         string
	CommitMessages string
	MinScore       *int
	MaxScore       *int
	StartTime      *time.Time
	EndTime        *time.Time
	Page           int
	Size           int
}

type PushReviewLogPage struct {
	Items []PushReviewLog `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

type MergeRequestReviewLogPage struct {
	Items []MergeRequestReviewLog `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}

type ReviewLogRepository interface {
	CreatePush(ctx context.Context, input PushReviewLogInput) (*PushReviewLog, error)
	FindPushByID(ctx context.Context, id uint) (*PushReviewLog, error)
	SearchPush(ctx context.Context, query ReviewLogSearchQuery) (*PushReviewLogPage, error)
	CreateMergeRequest(ctx context.Context, input MergeRequestReviewLogInput) (*MergeRequestReviewLog, error)
	FindMergeRequestByID(ctx context.Context, id uint) (*MergeRequestReviewLog, error)
	SearchMergeRequest(ctx context.Context, query ReviewLogSearchQuery) (*MergeRequestReviewLogPage, error)
}

type ReviewLogService struct {
	logs ReviewLogRepository
}

func NewReviewLogService(logs ReviewLogRepository) *ReviewLogService {
	return &ReviewLogService{logs: logs}
}

func (s *ReviewLogService) CreatePush(ctx context.Context, input PushReviewLogInput) (*PushReviewLog, error) {
	return s.logs.CreatePush(ctx, normalizePushReviewLogInput(input))
}

func (s *ReviewLogService) GetPush(ctx context.Context, id uint) (*PushReviewLog, error) {
	if id == 0 {
		return nil, ErrInvalidReviewLogInput
	}
	return s.logs.FindPushByID(ctx, id)
}

func (s *ReviewLogService) SearchPush(ctx context.Context, query ReviewLogSearchQuery) (*PushReviewLogPage, error) {
	return s.logs.SearchPush(ctx, normalizeReviewLogSearchQuery(query))
}

func (s *ReviewLogService) CreateMergeRequest(ctx context.Context, input MergeRequestReviewLogInput) (*MergeRequestReviewLog, error) {
	return s.logs.CreateMergeRequest(ctx, normalizeMergeRequestReviewLogInput(input))
}

func (s *ReviewLogService) GetMergeRequest(ctx context.Context, id uint) (*MergeRequestReviewLog, error) {
	if id == 0 {
		return nil, ErrInvalidReviewLogInput
	}
	return s.logs.FindMergeRequestByID(ctx, id)
}

func (s *ReviewLogService) SearchMergeRequest(ctx context.Context, query ReviewLogSearchQuery) (*MergeRequestReviewLogPage, error) {
	return s.logs.SearchMergeRequest(ctx, normalizeReviewLogSearchQuery(query))
}

func normalizePushReviewLogInput(input PushReviewLogInput) PushReviewLogInput {
	input.ProjectName = strings.TrimSpace(input.ProjectName)
	input.Author = strings.TrimSpace(input.Author)
	input.AuthorIdentity = strings.TrimSpace(input.AuthorIdentity)
	input.AuthorDisplayName = strings.TrimSpace(input.AuthorDisplayName)
	input.Branch = strings.TrimSpace(input.Branch)
	input.LastCommitURL = strings.TrimSpace(input.LastCommitURL)
	if input.Author == "" {
		input.Author = input.AuthorIdentity
	}
	return input
}

func normalizeMergeRequestReviewLogInput(input MergeRequestReviewLogInput) MergeRequestReviewLogInput {
	input.ProjectName = strings.TrimSpace(input.ProjectName)
	input.Author = strings.TrimSpace(input.Author)
	input.AuthorIdentity = strings.TrimSpace(input.AuthorIdentity)
	input.AuthorDisplayName = strings.TrimSpace(input.AuthorDisplayName)
	input.SourceBranch = strings.TrimSpace(input.SourceBranch)
	input.TargetBranch = strings.TrimSpace(input.TargetBranch)
	input.LastCommitID = strings.TrimSpace(input.LastCommitID)
	input.URL = strings.TrimSpace(input.URL)
	if input.Author == "" {
		input.Author = input.AuthorIdentity
	}
	return input
}

func normalizeReviewLogSearchQuery(query ReviewLogSearchQuery) ReviewLogSearchQuery {
	query.Authors = cleanStringSlice(query.Authors)
	query.ProjectNames = cleanStringSlice(query.ProjectNames)
	query.Branch = strings.TrimSpace(query.Branch)
	query.CommitMessages = strings.TrimSpace(query.CommitMessages)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}
	if query.Size > 200 {
		query.Size = 200
	}
	return query
}

func authorDisplayText(identity, displayName string) string {
	identity = strings.TrimSpace(identity)
	displayName = strings.TrimSpace(displayName)
	switch {
	case identity == "":
		return displayName
	case displayName == "":
		return identity
	default:
		return identity + "（" + displayName + "）"
	}
}
