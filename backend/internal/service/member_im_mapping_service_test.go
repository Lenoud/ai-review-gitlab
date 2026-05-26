package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemberIMMappingServiceCreateNormalizesInputAndRejectsDuplicates(t *testing.T) {
	repo := &fakeMemberIMMappingRepository{
		exists: func(ctx context.Context, gitUsername, platform string, excludeID uint) (bool, error) {
			require.Equal(t, "alice", gitUsername)
			require.Equal(t, IMRobotPlatformDingTalk, platform)
			require.Zero(t, excludeID)
			return false, nil
		},
		create: func(ctx context.Context, input MemberIMMappingInput) (*MemberIMMapping, error) {
			require.Equal(t, "alice", input.GitUsername)
			require.Equal(t, IMRobotPlatformDingTalk, input.Platform)
			require.Equal(t, "ding-user-1", input.IMUserID)
			require.Equal(t, "Alice Chen", input.DisplayName)
			return &MemberIMMapping{ID: 1, GitUsername: input.GitUsername, Platform: input.Platform, IMUserID: input.IMUserID}, nil
		},
	}
	svc := NewMemberIMMappingService(repo)

	mapping, err := svc.Create(context.Background(), MemberIMMappingInput{
		GitUsername: " alice ",
		Platform:    " dingtalk ",
		IMUserID:    " ding-user-1 ",
		DisplayName: " Alice Chen ",
	})

	require.NoError(t, err)
	require.Equal(t, uint(1), mapping.ID)
	require.Equal(t, "alice", mapping.GitUsername)
}

func TestMemberIMMappingServiceCreateRejectsDuplicateGitUsernameAndPlatform(t *testing.T) {
	repo := &fakeMemberIMMappingRepository{
		exists: func(ctx context.Context, gitUsername, platform string, excludeID uint) (bool, error) {
			return true, nil
		},
	}
	svc := NewMemberIMMappingService(repo)

	_, err := svc.Create(context.Background(), MemberIMMappingInput{
		GitUsername: "alice",
		Platform:    IMRobotPlatformFeishu,
		IMUserID:    "ou_123",
	})

	require.ErrorIs(t, err, ErrMemberIMMappingExists)
}

func TestMemberIMMappingServiceRejectsInvalidInput(t *testing.T) {
	svc := NewMemberIMMappingService(&fakeMemberIMMappingRepository{})

	tests := []MemberIMMappingInput{
		{Platform: IMRobotPlatformDingTalk, IMUserID: "ding-user"},
		{GitUsername: "alice", IMUserID: "ding-user"},
		{GitUsername: "alice", Platform: "slack", IMUserID: "slack-user"},
		{GitUsername: "alice", Platform: IMRobotPlatformDingTalk},
	}

	for _, input := range tests {
		_, err := svc.Create(context.Background(), input)
		require.ErrorIs(t, err, ErrInvalidMemberIMMappingInput)
	}
}

func TestMemberIMMappingServiceRejectsOverlongInput(t *testing.T) {
	svc := NewMemberIMMappingService(&fakeMemberIMMappingRepository{})

	tests := []MemberIMMappingInput{
		{GitUsername: strings.Repeat("a", 129), Platform: IMRobotPlatformDingTalk, IMUserID: "ding-user"},
		{GitUsername: "alice", Platform: IMRobotPlatformDingTalk, IMUserID: strings.Repeat("u", 257)},
		{GitUsername: "alice", Platform: IMRobotPlatformDingTalk, IMUserID: "ding-user", DisplayName: strings.Repeat("d", 129)},
	}

	for _, input := range tests {
		_, err := svc.Create(context.Background(), input)
		require.ErrorIs(t, err, ErrInvalidMemberIMMappingInput)
	}
}

func TestMemberIMMappingServiceUpdateRequiresIDAndExcludesItFromDuplicateCheck(t *testing.T) {
	repo := &fakeMemberIMMappingRepository{
		exists: func(ctx context.Context, gitUsername, platform string, excludeID uint) (bool, error) {
			require.Equal(t, uint(9), excludeID)
			return false, nil
		},
		update: func(ctx context.Context, id uint, input MemberIMMappingInput) (*MemberIMMapping, error) {
			require.Equal(t, uint(9), id)
			return &MemberIMMapping{ID: id, GitUsername: input.GitUsername, Platform: input.Platform, IMUserID: input.IMUserID}, nil
		},
	}
	svc := NewMemberIMMappingService(repo)

	mapping, err := svc.Update(context.Background(), 9, MemberIMMappingInput{
		GitUsername: "alice",
		Platform:    IMRobotPlatformWeCom,
		IMUserID:    "wecom-user",
	})

	require.NoError(t, err)
	require.Equal(t, uint(9), mapping.ID)
}

func TestMemberIMMappingServiceUpdateRequiresNonZeroID(t *testing.T) {
	svc := NewMemberIMMappingService(&fakeMemberIMMappingRepository{})

	_, err := svc.Update(context.Background(), 0, MemberIMMappingInput{
		GitUsername: "alice",
		Platform:    IMRobotPlatformDingTalk,
		IMUserID:    "ding-user",
	})

	require.ErrorIs(t, err, ErrInvalidMemberIMMappingInput)
}

func TestMemberIMMappingServiceDeleteCleansIDs(t *testing.T) {
	repo := &fakeMemberIMMappingRepository{
		delete: func(ctx context.Context, ids []uint) error {
			require.Equal(t, []uint{3, 2}, ids)
			return nil
		},
	}
	svc := NewMemberIMMappingService(repo)

	err := svc.Delete(context.Background(), []uint{0, 3, 3, 2})

	require.NoError(t, err)
}

func TestMemberIMMappingServiceDeleteRejectsEmptyIDs(t *testing.T) {
	svc := NewMemberIMMappingService(&fakeMemberIMMappingRepository{})

	err := svc.Delete(context.Background(), []uint{0})

	require.ErrorIs(t, err, ErrInvalidMemberIMMappingInput)
}

func TestMemberIMMappingServiceSearchNormalizesQuery(t *testing.T) {
	repo := &fakeMemberIMMappingRepository{
		search: func(ctx context.Context, query MemberIMMappingSearchQuery) (*MemberIMMappingPage, error) {
			require.Equal(t, "alice", query.Keyword)
			require.Equal(t, IMRobotPlatformFeishu, query.Platform)
			require.Equal(t, 1, query.Page)
			require.Equal(t, 20, query.Size)
			return &MemberIMMappingPage{Items: []MemberIMMapping{{ID: 1, GitUsername: "alice"}}, Total: 1, Page: 1, Size: 20}, nil
		},
	}
	svc := NewMemberIMMappingService(repo)

	page, err := svc.Search(context.Background(), MemberIMMappingSearchQuery{
		Keyword:  " alice ",
		Platform: " feishu ",
		Page:     -1,
		Size:     0,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
}

type fakeMemberIMMappingRepository struct {
	create func(context.Context, MemberIMMappingInput) (*MemberIMMapping, error)
	update func(context.Context, uint, MemberIMMappingInput) (*MemberIMMapping, error)
	find   func(context.Context, uint) (*MemberIMMapping, error)
	delete func(context.Context, []uint) error
	search func(context.Context, MemberIMMappingSearchQuery) (*MemberIMMappingPage, error)
	exists func(context.Context, string, string, uint) (bool, error)
}

func (r *fakeMemberIMMappingRepository) CreateMemberIMMapping(ctx context.Context, input MemberIMMappingInput) (*MemberIMMapping, error) {
	if r.create == nil {
		return nil, errors.New("unexpected create")
	}
	return r.create(ctx, input)
}

func (r *fakeMemberIMMappingRepository) UpdateMemberIMMapping(ctx context.Context, id uint, input MemberIMMappingInput) (*MemberIMMapping, error) {
	if r.update == nil {
		return nil, errors.New("unexpected update")
	}
	return r.update(ctx, id, input)
}

func (r *fakeMemberIMMappingRepository) FindMemberIMMappingByID(ctx context.Context, id uint) (*MemberIMMapping, error) {
	if r.find == nil {
		return nil, errors.New("unexpected find")
	}
	return r.find(ctx, id)
}

func (r *fakeMemberIMMappingRepository) DeleteMemberIMMappings(ctx context.Context, ids []uint) error {
	if r.delete == nil {
		return errors.New("unexpected delete")
	}
	return r.delete(ctx, ids)
}

func (r *fakeMemberIMMappingRepository) SearchMemberIMMappings(ctx context.Context, query MemberIMMappingSearchQuery) (*MemberIMMappingPage, error) {
	if r.search == nil {
		return nil, errors.New("unexpected search")
	}
	return r.search(ctx, query)
}

func (r *fakeMemberIMMappingRepository) ExistsMemberIMMapping(ctx context.Context, gitUsername string, platform string, excludeID uint) (bool, error) {
	if r.exists == nil {
		return false, errors.New("unexpected exists")
	}
	return r.exists(ctx, gitUsername, platform, excludeID)
}
