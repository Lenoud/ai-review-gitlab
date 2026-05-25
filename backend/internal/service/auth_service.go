package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	UserStatusEnabled  = "enabled"
	UserStatusDisabled = "disabled"

	tokenUseAccess  = "access"
	tokenUseRefresh = "refresh"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserDisabled       = errors.New("user disabled")
	ErrUserNotFound       = errors.New("user not found")
)

type User struct {
	ID           uint
	Username     string
	PasswordHash string
	Nickname     string
	Status       string
}

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
}

type LoginInput struct {
	Username string
	Password string
}

type TokenPair struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
}

type AuthSubject struct {
	UserID   uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type AuthOptions struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
	Now             func() time.Time
}

type AuthService struct {
	users           UserRepository
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
	now             func() time.Time
}

func NewAuthService(users UserRepository, opts AuthOptions) (*AuthService, error) {
	if users == nil {
		return nil, errors.New("user repository is required")
	}
	if strings.TrimSpace(opts.JWTSecret) == "" {
		return nil, errors.New("jwt secret is required")
	}
	if opts.AccessTokenTTL <= 0 {
		return nil, errors.New("access token ttl must be positive")
	}
	if opts.RefreshTokenTTL <= 0 {
		return nil, errors.New("refresh token ttl must be positive")
	}
	if strings.TrimSpace(opts.Issuer) == "" {
		opts.Issuer = "ai-review"
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return &AuthService{
		users:           users,
		secret:          []byte(opts.JWTSecret),
		accessTokenTTL:  opts.AccessTokenTTL,
		refreshTokenTTL: opts.RefreshTokenTTL,
		issuer:          opts.Issuer,
		now:             opts.Now,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	username := strings.TrimSpace(input.Username)
	if username == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.Status != "" && user.Status != UserStatusEnabled {
		return nil, ErrUserDisabled
	}
	if !CheckPassword(user.PasswordHash, input.Password) {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokenPair(user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	subject, err := s.validateToken(refreshToken, tokenUseRefresh)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(ctx, subject.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if user.Status != "" && user.Status != UserStatusEnabled {
		return nil, ErrUserDisabled
	}
	return s.issueTokenPair(user)
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, token string) (*AuthSubject, error) {
	subject, err := s.validateToken(token, tokenUseAccess)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(ctx, subject.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if user.Status != "" && user.Status != UserStatusEnabled {
		return nil, ErrUserDisabled
	}
	subject.Nickname = user.Nickname
	return subject, nil
}

func (s *AuthService) issueTokenPair(user *User) (*TokenPair, error) {
	access, err := s.issueToken(user, tokenUseAccess, s.accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := s.issueToken(user, tokenUseRefresh, s.refreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:      access,
		RefreshToken:     refresh,
		TokenType:        "Bearer",
		ExpiresIn:        int64(s.accessTokenTTL.Seconds()),
		RefreshExpiresIn: int64(s.refreshTokenTTL.Seconds()),
	}, nil
}

func (s *AuthService) issueToken(user *User, tokenUse string, ttl time.Duration) (string, error) {
	now := s.now().UTC()
	claims := authClaims{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		TokenUse: tokenUse,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *AuthService) validateToken(rawToken string, expectedUse string) (*AuthSubject, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, ErrInvalidToken
	}

	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer))
	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokenUse != expectedUse || claims.UserID == 0 || strings.TrimSpace(claims.Username) == "" {
		return nil, ErrInvalidToken
	}
	return &AuthSubject{
		UserID:   claims.UserID,
		Username: claims.Username,
		Nickname: claims.Nickname,
	}, nil
}

type authClaims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
	TokenUse string `json:"tokenUse"`
	jwt.RegisteredClaims
}
