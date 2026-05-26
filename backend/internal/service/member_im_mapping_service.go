package service

import (
	"context"
	"errors"
	"strings"
)

const (
	memberIMMappingGitUsernameMaxLength = 128
	memberIMMappingPlatformMaxLength    = 32
	memberIMMappingIMUserIDMaxLength    = 256
	memberIMMappingDisplayNameMaxLength = 128
)

var (
	ErrMemberIMMappingNotFound     = errors.New("member im mapping not found")
	ErrInvalidMemberIMMappingInput = errors.New("invalid member im mapping input")
	ErrMemberIMMappingExists       = errors.New("member im mapping exists")
)

type MemberIMMapping struct {
	ID          uint   `json:"id"`
	GitUsername string `json:"gitUsername"`
	Platform    string `json:"platform"`
	IMUserID    string `json:"imUserId"`
	DisplayName string `json:"displayName"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type MemberIMMappingInput struct {
	GitUsername string
	Platform    string
	IMUserID    string
	DisplayName string
}

type MemberIMMappingSearchQuery struct {
	Keyword  string
	Platform string
	Page     int
	Size     int
}

type MemberIMMappingPage struct {
	Items []MemberIMMapping `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

type MemberIMMappingRepository interface {
	CreateMemberIMMapping(ctx context.Context, input MemberIMMappingInput) (*MemberIMMapping, error)
	UpdateMemberIMMapping(ctx context.Context, id uint, input MemberIMMappingInput) (*MemberIMMapping, error)
	FindMemberIMMappingByID(ctx context.Context, id uint) (*MemberIMMapping, error)
	DeleteMemberIMMappings(ctx context.Context, ids []uint) error
	SearchMemberIMMappings(ctx context.Context, query MemberIMMappingSearchQuery) (*MemberIMMappingPage, error)
	ExistsMemberIMMapping(ctx context.Context, gitUsername string, platform string, excludeID uint) (bool, error)
}

type MemberIMMappingService struct {
	mappings MemberIMMappingRepository
}

func NewMemberIMMappingService(mappings MemberIMMappingRepository) *MemberIMMappingService {
	return &MemberIMMappingService{mappings: mappings}
}

func (s *MemberIMMappingService) Create(ctx context.Context, input MemberIMMappingInput) (*MemberIMMapping, error) {
	normalized, err := normalizeMemberIMMappingInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.mappings.ExistsMemberIMMapping(ctx, normalized.GitUsername, normalized.Platform, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrMemberIMMappingExists
	}
	return s.mappings.CreateMemberIMMapping(ctx, normalized)
}

func (s *MemberIMMappingService) Update(ctx context.Context, id uint, input MemberIMMappingInput) (*MemberIMMapping, error) {
	if id == 0 {
		return nil, ErrInvalidMemberIMMappingInput
	}
	normalized, err := normalizeMemberIMMappingInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.mappings.ExistsMemberIMMapping(ctx, normalized.GitUsername, normalized.Platform, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrMemberIMMappingExists
	}
	return s.mappings.UpdateMemberIMMapping(ctx, id, normalized)
}

func (s *MemberIMMappingService) Get(ctx context.Context, id uint) (*MemberIMMapping, error) {
	if id == 0 {
		return nil, ErrInvalidMemberIMMappingInput
	}
	return s.mappings.FindMemberIMMappingByID(ctx, id)
}

func (s *MemberIMMappingService) Delete(ctx context.Context, ids []uint) error {
	cleanIDs := make([]uint, 0, len(ids))
	seen := map[uint]struct{}{}
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		cleanIDs = append(cleanIDs, id)
	}
	if len(cleanIDs) == 0 {
		return ErrInvalidMemberIMMappingInput
	}
	return s.mappings.DeleteMemberIMMappings(ctx, cleanIDs)
}

func (s *MemberIMMappingService) Search(ctx context.Context, query MemberIMMappingSearchQuery) (*MemberIMMappingPage, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Platform = strings.TrimSpace(query.Platform)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}
	if query.Size > 200 {
		query.Size = 200
	}
	return s.mappings.SearchMemberIMMappings(ctx, query)
}

func normalizeMemberIMMappingInput(input MemberIMMappingInput) (MemberIMMappingInput, error) {
	input.GitUsername = strings.TrimSpace(input.GitUsername)
	input.Platform = strings.TrimSpace(input.Platform)
	input.IMUserID = strings.TrimSpace(input.IMUserID)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.GitUsername == "" || input.Platform == "" || input.IMUserID == "" {
		return MemberIMMappingInput{}, ErrInvalidMemberIMMappingInput
	}
	if len(input.GitUsername) > memberIMMappingGitUsernameMaxLength ||
		len(input.Platform) > memberIMMappingPlatformMaxLength ||
		len(input.IMUserID) > memberIMMappingIMUserIDMaxLength ||
		len(input.DisplayName) > memberIMMappingDisplayNameMaxLength {
		return MemberIMMappingInput{}, ErrInvalidMemberIMMappingInput
	}
	if !isSupportedIMRobotPlatform(input.Platform) {
		return MemberIMMappingInput{}, ErrInvalidMemberIMMappingInput
	}
	return input, nil
}
