package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMemberIMMappingHandlerCreateUpdateGetSearchAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewMemberIMMappingHandler(&fakeMemberIMMappingService{
		create: func(ctx context.Context, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
			require.Equal(t, "alice", input.GitUsername)
			require.Equal(t, service.IMRobotPlatformDingTalk, input.Platform)
			require.Equal(t, "ding-user", input.IMUserID)
			return &service.MemberIMMapping{ID: 11, GitUsername: input.GitUsername, Platform: input.Platform, IMUserID: input.IMUserID}, nil
		},
		update: func(ctx context.Context, id uint, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
			require.Equal(t, uint(11), id)
			require.Equal(t, service.IMRobotPlatformFeishu, input.Platform)
			return &service.MemberIMMapping{ID: id, GitUsername: input.GitUsername, Platform: input.Platform, IMUserID: input.IMUserID}, nil
		},
		get: func(ctx context.Context, id uint) (*service.MemberIMMapping, error) {
			require.Equal(t, uint(11), id)
			return &service.MemberIMMapping{ID: 11, GitUsername: "alice", Platform: service.IMRobotPlatformDingTalk, IMUserID: "ding-user"}, nil
		},
		search: func(ctx context.Context, query service.MemberIMMappingSearchQuery) (*service.MemberIMMappingPage, error) {
			require.Equal(t, "alice", query.Keyword)
			require.Equal(t, service.IMRobotPlatformDingTalk, query.Platform)
			require.Equal(t, 2, query.Page)
			require.Equal(t, 5, query.Size)
			return &service.MemberIMMappingPage{Items: []service.MemberIMMapping{{ID: 11, GitUsername: "alice"}}, Total: 1, Page: 2, Size: 5}, nil
		},
		delete: func(ctx context.Context, ids []uint) error {
			require.Equal(t, []uint{11}, ids)
			return nil
		},
	})
	r.POST("/create", handler.Create)
	r.POST("/update", handler.Update)
	r.GET("/get", handler.Get)
	r.GET("/search", handler.Search)
	r.POST("/delete", handler.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(`{"gitUsername":"alice","platform":"dingtalk","imUserId":"ding-user","displayName":"Alice"}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(`{"id":11,"gitUsername":"alice","platform":"feishu","imUserId":"ou_123"}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/get?id=11", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/search?keyword=alice&platform=dingtalk&page=2&size=5", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBufferString(`{"ids":[11]}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestMemberIMMappingHandlerMapsErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewMemberIMMappingHandler(&fakeMemberIMMappingService{
		get: func(ctx context.Context, id uint) (*service.MemberIMMapping, error) {
			return nil, service.ErrMemberIMMappingNotFound
		},
		create: func(ctx context.Context, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
			return nil, service.ErrMemberIMMappingExists
		},
	})
	r.GET("/get", handler.Get)
	r.POST("/create", handler.Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get?id=404", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(`{"gitUsername":"alice","platform":"dingtalk","imUserId":"ding-user"}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

type fakeMemberIMMappingService struct {
	create func(context.Context, service.MemberIMMappingInput) (*service.MemberIMMapping, error)
	update func(context.Context, uint, service.MemberIMMappingInput) (*service.MemberIMMapping, error)
	get    func(context.Context, uint) (*service.MemberIMMapping, error)
	delete func(context.Context, []uint) error
	search func(context.Context, service.MemberIMMappingSearchQuery) (*service.MemberIMMappingPage, error)
}

func (s *fakeMemberIMMappingService) Create(ctx context.Context, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
	return s.create(ctx, input)
}

func (s *fakeMemberIMMappingService) Update(ctx context.Context, id uint, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
	return s.update(ctx, id, input)
}

func (s *fakeMemberIMMappingService) Get(ctx context.Context, id uint) (*service.MemberIMMapping, error) {
	return s.get(ctx, id)
}

func (s *fakeMemberIMMappingService) Delete(ctx context.Context, ids []uint) error {
	return s.delete(ctx, ids)
}

func (s *fakeMemberIMMappingService) Search(ctx context.Context, query service.MemberIMMappingSearchQuery) (*service.MemberIMMappingPage, error) {
	return s.search(ctx, query)
}
