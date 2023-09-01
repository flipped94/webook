package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/flipped94/webook/internal/domain"
	"github.com/flipped94/webook/internal/repository"
)

var (
	ErrInvalidUserOrPassword = errors.New("邮箱或者密码不正确")
	ErrUserDuplicate         = repository.ErrUserDuplicate
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Edit(ctx context.Context, user domain.User) error
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type userService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) UserService {
	return &userService{
		repository: repository,
	}
}

func (us *userService) Signup(ctx context.Context, u domain.User) error {
	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 存储
	return us.repository.Create(ctx, u)
}

func (us *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := us.repository.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (us *userService) Edit(ctx context.Context, user domain.User) error {
	return us.repository.Edit(ctx, user)
}

func (us *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return us.repository.FindById(ctx, id)
}

func (us *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := us.repository.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	u = domain.User{
		Phone: phone,
	}
	err = us.repository.Create(ctx, u)
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	return us.repository.FindByPhone(ctx, phone)
}
