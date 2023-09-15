package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/flipped94/webook/internal/domain"
	"github.com/flipped94/webook/internal/repository/cache"
	"github.com/flipped94/webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrDataNotFound
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	Edit(ctx context.Context, u domain.User) error
	FindByWechat(ctx context.Context, openid string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDao, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM users WHERE email = ?
	u, err := r.dao.FindByEmail(ctx, email)
	return r.entityToDomain(u), err
}

func (r *CachedUserRepository) Edit(ctx context.Context, u domain.User) error {
	// UPDATE users set nickname = ? where id = ?
	return r.dao.Update(ctx, dao.User{
		Id:        u.Id,
		Nickname:  u.Nickname,
		Birthday:  u.Birthday,
		Biography: u.Biography,
	})

}

func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache
	u, err := r.cache.Get(ctx, id)
	switch err {
	case nil:
		return u, err
	case cache.ErrUserNotFound:
		// 从数据库
		ue, err := r.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		u = r.entityToDomain(ue)
		go func() {
			// 找到回写 cache
			err := r.cache.Set(ctx, u)
			if err != nil {
				log.Fatal(err)
			}
		}()
		return u, nil
	default:
		return domain.User{}, err
	}

}

func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	ue, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(ue), nil
}

func (r *CachedUserRepository) FindByWechat(ctx context.Context, openid string) (domain.User, error) {
	u, err := r.dao.FindByWechat(ctx, openid)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) entityToDomain(ue dao.User) domain.User {
	return domain.User{
		Id:        ue.Id,
		Email:     ue.Email.String,
		Phone:     ue.Phone.String,
		Password:  ue.Password,
		Nickname:  ue.Nickname,
		Birthday:  ue.Birthday,
		Biography: ue.Biography,
		Ctime:     ue.Ctime,
		WechatInfo: domain.WechatInfo{
			OpenID:  ue.WechatOpenID.String,
			UnionID: ue.WechatUnionID.String,
		},
	}
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		WechatOpenID: sql.NullString{
			String: u.WechatInfo.OpenID,
			Valid:  u.WechatInfo.OpenID != "",
		},
		WechatUnionID: sql.NullString{
			String: u.WechatInfo.UnionID,
			Valid:  u.WechatInfo.UnionID != "",
		},
		Password:  u.Password,
		Nickname:  u.Nickname,
		Birthday:  u.Birthday,
		Biography: u.Biography,
	}
}
