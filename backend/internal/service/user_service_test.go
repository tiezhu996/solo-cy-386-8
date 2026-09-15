package service

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"gorm.io/gorm"
)

// fakeUserRepo 内存版用户仓储（表驱动测试用）。
type fakeUserRepo struct {
	users map[uint]*model.User
	seq   uint
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[uint]*model.User{}}
}

func (f *fakeUserRepo) Create(u *model.User) error {
	f.seq++
	u.ID = f.seq
	u.CreatedAt = time.Now()
	f.users[u.ID] = u
	return nil
}
func (f *fakeUserRepo) GetByID(id uint) (*model.User, error) {
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeUserRepo) GetByUsername(username string) (*model.User, error) {
	for _, u := range f.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *fakeUserRepo) List(page, pageSize int) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) Update(u *model.User) error {
	if _, ok := f.users[u.ID]; !ok {
		return repository.ErrNotFound
	}
	f.users[u.ID] = u
	return nil
}
func (f *fakeUserRepo) CreateWithTx(tx *gorm.DB, u *model.User) error { return f.Create(u) }
func (f *fakeUserRepo) UpdateCredit(tx *gorm.DB, id uint, delta int) error {
	if u, ok := f.users[id]; ok {
		u.CreditScore += delta
		return nil
	}
	return repository.ErrNotFound
}

func newTestUserService() (*UserService, *fakeUserRepo) {
	repo := newFakeUserRepo()
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	svc := NewUserService(repo, logger, "test-secret", 72*time.Hour)
	return svc, repo
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestUserServiceRegister(t *testing.T) {
	svc, _ := newTestUserService()
	cases := []struct {
		name    string
		req     dto.RegisterRequest
		wantErr bool
	}{
		{"valid", dto.RegisterRequest{Username: "alice", Password: "123456", Nickname: "Alice"}, false},
		{"duplicate", dto.RegisterRequest{Username: "alice", Password: "123456", Nickname: "Alice"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(tc.req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Register() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserServiceLogin(t *testing.T) {
	svc, _ := newTestUserService()
	_, _ = svc.Register(dto.RegisterRequest{Username: "bob", Password: "123456", Nickname: "Bob"})
	cases := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{"correct", "bob", "123456", false},
		{"wrong password", "bob", "wrong", true},
		{"not exists", "nobody", "123456", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, token, err := svc.Login(dto.LoginRequest{Username: tc.username, Password: tc.password})
			if tc.wantErr {
				if err == nil {
					t.Fatal("Login() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Login() unexpected error: %v", err)
			}
			if token == "" {
				t.Fatal("Login() returned empty token")
			}
		})
	}
}

func TestUserServiceUpdateRole(t *testing.T) {
	svc, repo := newTestUserService()
	u, _ := svc.Register(dto.RegisterRequest{Username: "carol", Password: "123456", Nickname: "Carol"})
	cases := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{"to admin", "admin", false},
		{"invalid role", "superman", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.UpdateRole(1, u.ID, tc.role)
			if (err != nil) != tc.wantErr {
				t.Fatalf("UpdateRole() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
	if repo.users[u.ID].Role != "admin" {
		t.Fatalf("expected role admin, got %s", repo.users[u.ID].Role)
	}
}

var _ = errors.Is
