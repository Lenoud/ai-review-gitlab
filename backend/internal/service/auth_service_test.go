package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceLoginIssuesUsableAccessAndRefreshTokens(t *testing.T) {
	hash, err := HashPassword("correct-password")
	require.NoError(t, err)

	repo := newMemoryUserRepository(&User{
		ID:           1,
		Username:     "admin",
		PasswordHash: hash,
		Nickname:     "Administrator",
		Status:       UserStatusEnabled,
	})
	svc, err := NewAuthService(repo, AuthOptions{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
		Issuer:          "ai-review-test",
		Now:             time.Now,
	})
	require.NoError(t, err)

	tokens, err := svc.Login(context.Background(), LoginInput{
		Username: "admin",
		Password: "correct-password",
	})

	require.NoError(t, err)
	require.Equal(t, "Bearer", tokens.TokenType)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
	require.Equal(t, int64(60), tokens.ExpiresIn)
	require.Equal(t, int64(3600), tokens.RefreshExpiresIn)

	subject, err := svc.ValidateAccessToken(context.Background(), tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, uint(1), subject.UserID)
	require.Equal(t, "admin", subject.Username)
	require.Equal(t, "Administrator", subject.Nickname)

	refreshed, err := svc.Refresh(context.Background(), tokens.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, refreshed.AccessToken)

	_, err = svc.ValidateAccessToken(context.Background(), tokens.RefreshToken)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestAuthServiceRejectsInvalidCredentials(t *testing.T) {
	hash, err := HashPassword("correct-password")
	require.NoError(t, err)

	svc, err := NewAuthService(newMemoryUserRepository(&User{
		ID:           1,
		Username:     "admin",
		PasswordHash: hash,
		Status:       UserStatusEnabled,
	}), AuthOptions{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
		Issuer:          "ai-review-test",
	})
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), LoginInput{
		Username: "admin",
		Password: "wrong-password",
	})

	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestRBACServiceListsRolesAndPermissionGroups(t *testing.T) {
	repo := &memoryRBACRepository{
		roles: []Role{
			{ID: 1, Code: "admin", Name: "管理员"},
		},
		permissionGroups: []PermissionGroup{
			{
				Category: "project",
				Permissions: []Permission{
					{ID: 1, Code: "project:read", Name: "查看项目", Category: "project"},
				},
			},
		},
	}
	svc := NewRBACService(repo)

	roles, err := svc.ListRoles(context.Background())
	require.NoError(t, err)
	require.Equal(t, repo.roles, roles)

	groups, err := svc.ListPermissionGroups(context.Background())
	require.NoError(t, err)
	require.Equal(t, repo.permissionGroups, groups)
}

func TestRBACServiceCreateRoleNormalizesAndDeduplicatesPermissionIDs(t *testing.T) {
	repo := &memoryRBACRepository{}
	svc := NewRBACService(repo)

	role, err := svc.CreateRole(context.Background(), RoleInput{
		Code:          " reviewer ",
		Name:          " 审查员 ",
		Description:   " review projects ",
		PermissionIDs: []uint{2, 1, 2, 0},
	})

	require.NoError(t, err)
	require.Equal(t, "reviewer", repo.created.Code)
	require.Equal(t, "审查员", repo.created.Name)
	require.Equal(t, "review projects", repo.created.Description)
	require.Equal(t, []uint{2, 1}, repo.created.PermissionIDs)
	require.Equal(t, repo.roleDetail, role)
}

func TestRBACServiceRejectsInvalidRoleInput(t *testing.T) {
	svc := NewRBACService(&memoryRBACRepository{})

	_, err := svc.CreateRole(context.Background(), RoleInput{Name: "审查员"})
	require.ErrorIs(t, err, ErrInvalidRBACInput)

	_, err = svc.UpdateRole(context.Background(), 0, RoleInput{Code: "reviewer", Name: "审查员"})
	require.ErrorIs(t, err, ErrInvalidRBACInput)
}

func TestRBACServiceDeleteRoleRejectsEmptyIDs(t *testing.T) {
	svc := NewRBACService(&memoryRBACRepository{})

	err := svc.DeleteRoles(context.Background(), []uint{0, 0})

	require.ErrorIs(t, err, ErrInvalidRBACInput)
}

type memoryUserRepository struct {
	byID       map[uint]*User
	byUsername map[string]*User
}

func newMemoryUserRepository(users ...*User) *memoryUserRepository {
	repo := &memoryUserRepository{
		byID:       map[uint]*User{},
		byUsername: map[string]*User{},
	}
	for _, user := range users {
		repo.byID[user.ID] = user
		repo.byUsername[user.Username] = user
	}
	return repo
}

func (r *memoryUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	user, ok := r.byUsername[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *memoryUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	user, ok := r.byID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

type memoryRBACRepository struct {
	roles            []Role
	permissionGroups []PermissionGroup
	created          RoleInput
	updatedID        uint
	updated          RoleInput
	deletedIDs       []uint
	roleDetail       *RoleDetail
}

func (r *memoryRBACRepository) ListRoles(ctx context.Context) ([]Role, error) {
	return r.roles, nil
}

func (r *memoryRBACRepository) ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error) {
	return r.permissionGroups, nil
}

func (r *memoryRBACRepository) CreateRole(ctx context.Context, input RoleInput) (*RoleDetail, error) {
	r.created = input
	if r.roleDetail == nil {
		r.roleDetail = &RoleDetail{ID: 1, Code: input.Code, Name: input.Name, Description: input.Description, PermissionIDs: input.PermissionIDs}
	}
	return r.roleDetail, nil
}

func (r *memoryRBACRepository) UpdateRole(ctx context.Context, id uint, input RoleInput) (*RoleDetail, error) {
	r.updatedID = id
	r.updated = input
	if r.roleDetail == nil {
		r.roleDetail = &RoleDetail{ID: id, Code: input.Code, Name: input.Name, Description: input.Description, PermissionIDs: input.PermissionIDs}
	}
	return r.roleDetail, nil
}

func (r *memoryRBACRepository) FindRoleByID(ctx context.Context, id uint) (*RoleDetail, error) {
	if r.roleDetail == nil {
		return nil, ErrRoleNotFound
	}
	return r.roleDetail, nil
}

func (r *memoryRBACRepository) DeleteRoles(ctx context.Context, ids []uint) error {
	r.deletedIDs = ids
	return nil
}
