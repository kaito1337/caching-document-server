package service

import (
	"context"
	"crypto/rand"
	tokenStorage "document-server/internal/storage/token"
	userStorage "document-server/internal/storage/user"
	"encoding/hex"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	hasMin8Chars     = regexp.MustCompile(`.{8,}`)
	hasUpperAndLower = regexp.MustCompile(`[a-z].*[A-Z]|[A-Z].*[a-z]`)
	hasNumber        = regexp.MustCompile(`\d`)
	hasSymbol        = regexp.MustCompile(`[\W_]`)
)

type UserService struct {
	userStorage  UserStorage
	tokenStorage TokenStorage
	adminToken   string
}

func NewUserService(userStorage UserStorage, tokenStorage TokenStorage, adminToken string) *UserService {
	return &UserService{
		userStorage:  userStorage,
		tokenStorage: tokenStorage,
		adminToken:   adminToken,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, login, password, token string) error {
	if !hasMin8Chars.MatchString(password) || !hasUpperAndLower.MatchString(password) ||
		!hasNumber.MatchString(password) || !hasSymbol.MatchString(password) {
		return errors.New("invalid login or password")
	}

	validAdminToken := s.ValidateAdminToken(token)
	if !validAdminToken {
		return errors.New("invalid admin token")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userStorage.Create(ctx, userStorage.User{Login: login, PasswordHash: string(passwordHash)}); err != nil {
		return err
	}

	return nil
}

func (s *UserService) Authenticate(ctx context.Context, login, password string) (string, error) {
	user, err := s.userStorage.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", err
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	userToken := tokenStorage.UserToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.tokenStorage.Create(ctx, userToken); err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) Logout(ctx context.Context, token string) error {
	userToken, err := s.tokenStorage.GetByToken(ctx, token)
	if err != nil {
		return err
	}
	if userToken.UserID == uuid.Nil {
		return errors.New("invalid token")
	}
	return s.tokenStorage.Delete(ctx, token)
}

func (s *UserService) ValidateAdminToken(token string) bool {
	return token == s.adminToken && s.adminToken != ""
}

func (s *UserService) GetUserByToken(ctx context.Context, token string) (*userStorage.User, error) {
	userToken, err := s.tokenStorage.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(userToken.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	user, err := s.userStorage.GetUserByID(ctx, userToken.UserID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
