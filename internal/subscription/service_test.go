package subscription_test

import (
	"errors"
	"os"
	"testing"
	"weather-app/internal/database/models"
	"weather-app/internal/database/repository"
	"weather-app/internal/subscription"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mockMailService struct {
	Called bool
	Err    error
}

func (m *mockMailService) SendConfirmationMail(email, confirmURL, unsubscribeURL string) error {
	m.Called = true
	return m.Err
}

type mockUserRepo struct {
	GetByEmailFunc                           func(email string) (*models.User, error)
	CreateUserWithSubscriptionAndTokensFunc  func(email, city, frequency string, tokenTypes []string, gen func() (string, error)) (*repository.CreateUserWithSubscriptionAndTokensResult, error)
	UpdateUserConfirmationAndDeleteTokenFunc func(userID uuid.UUID, tokenID uuid.UUID) error
	DeleteUserWithTokensAndSubscriptionFunc  func(userID uuid.UUID) error
}

func (r *mockUserRepo) GetByEmail(email string) (*models.User, error) {
	return r.GetByEmailFunc(email)
}
func (r *mockUserRepo) CreateUserWithSubscriptionAndTokens(email, city, frequency string, tokenTypes []string, gen func() (string, error)) (*repository.CreateUserWithSubscriptionAndTokensResult, error) {
	return r.CreateUserWithSubscriptionAndTokensFunc(email, city, frequency, tokenTypes, gen)
}
func (r *mockUserRepo) UpdateUserConfirmationAndDeleteToken(userID uuid.UUID, tokenID uuid.UUID) error {
	return r.UpdateUserConfirmationAndDeleteTokenFunc(userID, tokenID)
}
func (r *mockUserRepo) DeleteUserWithTokensAndSubscription(userID uuid.UUID) error {
	return r.DeleteUserWithTokensAndSubscriptionFunc(userID)
}

type mockTokenRepo struct {
	GetTokenFunc func(value string) (*models.Token, error)
}

func (r *mockTokenRepo) GetToken(value string) (*models.Token, error) {
	return r.GetTokenFunc(value)
}

func TestSubscribe_Success(t *testing.T) {
	os.Setenv("BASE_URL", "https://test.com")

	userRepo := &mockUserRepo{
		GetByEmailFunc: func(email string) (*models.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
		CreateUserWithSubscriptionAndTokensFunc: func(email, city, frequency string, tokenTypes []string, gen func() (string, error)) (*repository.CreateUserWithSubscriptionAndTokensResult, error) {
			return &repository.CreateUserWithSubscriptionAndTokensResult{
				Tokens: map[string]*models.Token{
					models.TokenTypeConfirm:     {Value: "confirm-token"},
					models.TokenTypeUnsubscribe: {Value: "unsubscribe-token"},
				},
			}, nil
		},
	}

	mail := &mockMailService{}
	svc := subscription.NewSubscriptionService(userRepo, nil, mail)

	err := svc.Subscribe("test@example.com", "Kyiv", "daily")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mail.Called {
		t.Error("expected mail service to be called")
	}
}

func TestSubscribe_ExistingUser(t *testing.T) {
	userRepo := &mockUserRepo{
		GetByEmailFunc: func(email string) (*models.User, error) {
			return &models.User{}, nil
		},
	}

	svc := subscription.NewSubscriptionService(userRepo, nil, &mockMailService{})

	err := svc.Subscribe("test@example.com", "Kyiv", "daily")
	if err != subscription.ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestSubscribe_MailError(t *testing.T) {
	userRepo := &mockUserRepo{
		GetByEmailFunc: func(email string) (*models.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
		CreateUserWithSubscriptionAndTokensFunc: func(email, city, frequency string, tokenTypes []string, gen func() (string, error)) (*repository.CreateUserWithSubscriptionAndTokensResult, error) {
			return &repository.CreateUserWithSubscriptionAndTokensResult{
				Tokens: map[string]*models.Token{
					models.TokenTypeConfirm:     {Value: "c"},
					models.TokenTypeUnsubscribe: {Value: "u"},
				},
			}, nil
		},
	}

	mail := &mockMailService{Err: errors.New("mail error")}
	svc := subscription.NewSubscriptionService(userRepo, nil, mail)

	err := svc.Subscribe("test@example.com", "Kyiv", "daily")
	if err == nil || !errors.Is(err, subscription.ErrConfirmationMailError) {
		t.Errorf("expected confirmation mail error, got %v", err)
	}
}

func TestConfirm_Success(t *testing.T) {
	token := &models.Token{
		Value:  "token123",
		Type:   models.TokenTypeConfirm,
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	userRepo := &mockUserRepo{
		UpdateUserConfirmationAndDeleteTokenFunc: func(userID uuid.UUID, tokenID uuid.UUID) error {
			return nil
		},
	}

	tokenRepo := &mockTokenRepo{
		GetTokenFunc: func(value string) (*models.Token, error) {
			return token, nil
		},
	}

	svc := subscription.NewSubscriptionService(userRepo, tokenRepo, nil)
	err := svc.Confirm("token123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestConfirm_WrongType(t *testing.T) {
	tokenRepo := &mockTokenRepo{
		GetTokenFunc: func(value string) (*models.Token, error) {
			return &models.Token{Type: models.TokenTypeUnsubscribe}, nil
		},
	}

	svc := subscription.NewSubscriptionService(nil, tokenRepo, nil)
	err := svc.Confirm("abc")
	if err != subscription.ErrTokenWrongType {
		t.Errorf("expected ErrTokenWrongType, got %v", err)
	}
}

func TestConfirm_TokenEmpty(t *testing.T) {
	svc := subscription.NewSubscriptionService(nil, nil, nil)

	err := svc.Confirm("")
	if err != subscription.ErrTokenEmpty {
		t.Errorf("expected ErrTokenEmpty, got %v", err)
	}
}
func TestConfirm_TokenNotFound(t *testing.T) {
	tokenRepo := &mockTokenRepo{
		GetTokenFunc: func(value string) (*models.Token, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	svc := subscription.NewSubscriptionService(nil, tokenRepo, nil)

	err := svc.Confirm("nonexistent-token")
	if err != subscription.ErrTokenNotFound {
		t.Errorf("expected ErrTokenNotFound, got %v", err)
	}
}

func TestConfirm_TokenDBError(t *testing.T) {
	expectedDBErr := errors.New("db connection timeout")

	tokenRepo := &mockTokenRepo{
		GetTokenFunc: func(value string) (*models.Token, error) {
			return nil, expectedDBErr
		},
	}

	svc := subscription.NewSubscriptionService(nil, tokenRepo, nil)

	err := svc.Confirm("token123")
	if err == nil || !errors.Is(err, expectedDBErr) {
		t.Errorf("expected wrapped db error, got %v", err)
	}
}

func TestUnsubscribe_Success(t *testing.T) {
	token := &models.Token{
		Type:   models.TokenTypeUnsubscribe,
		UserID: uuid.New(),
	}

	userRepo := &mockUserRepo{
		DeleteUserWithTokensAndSubscriptionFunc: func(userID uuid.UUID) error {
			return nil
		},
	}

	tokenRepo := &mockTokenRepo{
		GetTokenFunc: func(value string) (*models.Token, error) {
			return token, nil
		},
	}

	svc := subscription.NewSubscriptionService(userRepo, tokenRepo, nil)
	err := svc.Unsubscribe("abc")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUnsubscribe_TokenEmpty(t *testing.T) {
	svc := subscription.NewSubscriptionService(nil, nil, nil)

	err := svc.Unsubscribe("")
	if err != subscription.ErrTokenEmpty {
		t.Errorf("expected ErrTokenEmpty, got %v", err)
	}
}

func TestUnsubscribe_TokenNotFound(t *testing.T) {
	tokenRepo := &mockTokenRepo{
		GetTokenFunc: func(value string) (*models.Token, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	svc := subscription.NewSubscriptionService(nil, tokenRepo, nil)

	err := svc.Unsubscribe("nonexistent-token")
	if err != subscription.ErrTokenNotFound {
		t.Errorf("expected ErrTokenNotFound, got %v", err)
	}
}

func TestUnsubscribe_TokenDBError(t *testing.T) {
	expectedDBErr := errors.New("db connection timeout")

	tokenRepo := &mockTokenRepo{
		GetTokenFunc: func(value string) (*models.Token, error) {
			return nil, expectedDBErr
		},
	}

	svc := subscription.NewSubscriptionService(nil, tokenRepo, nil)

	err := svc.Unsubscribe("token123")
	if err == nil || !errors.Is(err, expectedDBErr) {
		t.Errorf("expected wrapped db error, got %v", err)
	}
}

func TestUnsubscribe_WrongType(t *testing.T) {
	tokenRepo := &mockTokenRepo{
		GetTokenFunc: func(value string) (*models.Token, error) {
			return &models.Token{Type: models.TokenTypeConfirm}, nil
		},
	}
	userRepo := &mockUserRepo{
		DeleteUserWithTokensAndSubscriptionFunc: func(userID uuid.UUID) error {
			return nil
		},
	}

	svc := subscription.NewSubscriptionService(userRepo, tokenRepo, nil)
	err := svc.Unsubscribe("abc")
	if err != subscription.ErrTokenWrongType {
		t.Errorf("expected ErrTokenWrongType, got %v", err)
	}
}
