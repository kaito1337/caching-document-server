package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	tokenStorage "document-server/internal/storage/token"
	userStorage "document-server/internal/storage/user"
	"encoding/hex"
	"errors"
	"log/slog"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/hedhyw/semerr/pkg/v1/semerr"
	"golang.org/x/crypto/bcrypt"
)

var (
	hasValidLogin    = regexp.MustCompile(`^[a-zA-Z0-9]{8,}$`)
	hasMin8Chars     = regexp.MustCompile(`.{8,}`)
	hasUpperAndLower = regexp.MustCompile(`[a-z].*[A-Z]|[A-Z].*[a-z]`)
	hasNumber        = regexp.MustCompile(`\d`)
	hasSymbol        = regexp.MustCompile(`[\W_]`)
)

type UserService struct {
	userStorage  UserStorage
	tokenStorage TokenStorage
	logger       *slog.Logger
	adminToken   string
}

func NewUserService(userStorage UserStorage, tokenStorage TokenStorage, logger *slog.Logger, adminToken string) *UserService {
	return &UserService{
		userStorage:  userStorage,
		tokenStorage: tokenStorage,
		adminToken:   adminToken,
		logger:       logger,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, login, password, token string) error {
	if !hasValidLogin.MatchString(login) {
		return semerr.NewBadRequestError(errors.New("login must be at least 8 characters, only Latin letters and digits"))
	}

	if !hasMin8Chars.MatchString(password) ||
		!hasUpperAndLower.MatchString(password) ||
		!hasNumber.MatchString(password) ||
		!hasSymbol.MatchString(password) {
		return semerr.NewBadRequestError(errors.New("password must be at least 8 characters, with upper and lower case letters, a number, and a symbol"))
	}

	if !s.ValidateAdminToken(token) {
		return semerr.NewUnauthorizedError(errors.New("invalid admin token"))
	}

	_, err := s.userStorage.GetUserByLogin(ctx, login)
	if err == nil {
		return semerr.NewBadRequestError(errors.New("user with this login already exists"))
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("failed to query user", slog.String("error", err.Error()))
		return semerr.NewInternalServerError(err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", slog.String("error", err.Error()))
		return semerr.NewInternalServerError(err)
	}

	newUser := userStorage.User{Login: login, PasswordHash: string(passwordHash)}
	if err := s.userStorage.Create(ctx, newUser); err != nil {
		s.logger.Error("failed to create user", slog.String("error", err.Error()))
		return semerr.NewInternalServerError(err)
	}

	s.logger.Info("user successfully registered", slog.String("login", login))
	return nil
}

func (s *UserService) Authenticate(ctx context.Context, login, password string) (string, error) {
	user, err := s.userStorage.GetUserByLogin(ctx, login)
	if err != nil {
		s.logger.Warn("authentication failed: user not found", slog.String("login", login))
		return "", semerr.NewBadRequestError(errors.New("invalid credentials"))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("authentication failed: incorrect password", slog.String("login", login))
		return "", semerr.NewBadRequestError(errors.New("invalid credentials"))
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.logger.Error("failed to generate token", slog.String("error", err.Error()))
		return "", semerr.NewInternalServerError(err)
	}
	token := hex.EncodeToString(tokenBytes)

	userToken := tokenStorage.UserToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.tokenStorage.Create(ctx, userToken); err != nil {
		s.logger.Error("failed to store token", slog.String("error", err.Error()))
		return "", semerr.NewInternalServerError(err)
	}

	s.logger.Info("user authenticated", slog.String("login", login), slog.String("user_id", user.ID.String()))
	return token, nil
}

func (s *UserService) Logout(ctx context.Context, token string) error {
	userToken, err := s.tokenStorage.GetByToken(ctx, token)
	if err != nil {
		s.logger.Warn("logout failed: token not found", slog.String("token", token))
		return semerr.NewBadRequestError(errors.New("invalid token"))
	}

	if userToken.UserID == uuid.Nil {
		s.logger.Warn("logout failed: invalid token", slog.String("token", token))
		return semerr.NewBadRequestError(errors.New("invalid token"))
	}

	if err := s.tokenStorage.Delete(ctx, token); err != nil {
		s.logger.Error("failed to delete token", slog.String("token", token), slog.String("error", err.Error()))
		return semerr.NewInternalServerError(err)
	}

	s.logger.Info("user logged out", slog.String("user_id", userToken.UserID.String()))
	return nil
}

func (s *UserService) ValidateAdminToken(token string) bool {
	return token == s.adminToken && s.adminToken != ""
}
