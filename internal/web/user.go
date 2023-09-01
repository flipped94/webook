package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/flipped94/webook/internal/domain"
	"github.com/flipped94/webook/internal/service"
)

const (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
	birthdayRegexPattern = "^\\d{4}-\\d{2}-\\d{2}$"

	userIdKey = "userId"

	loginSmsTemplate = "SMS_154950909"

	loginBiz = "login"
)

// 用户有关路由
type UserHandler struct {
	service          service.UserService
	codeSvc          service.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	birthdayRegexExp *regexp.Regexp
}

func NewUserHandler(service service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		service:          service,
		codeSvc:          codeSvc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		birthdayRegexExp: regexp.MustCompile(birthdayRegexPattern, regexp.None),
	}
}

func (u *UserHandler) RegisterRoutes(ctx *gin.Engine) {
	ug := ctx.Group("/users")
	ug.POST("/signup", u.Signup)
	// ug.POST("/login", u.Login)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)
}

func (u *UserHandler) Signup(ctx *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}
	var req SignupReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := u.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !isEmail {
		ctx.JSON(http.StatusOK, Result{
			Code: 3,
			Msg:  "邮箱不正确",
		})
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 3,
			Msg:  "两次输入的密码不相同",
		})
		return
	}
	isPassword, err := u.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !isPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 3,
			Msg:  "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
		})
		return
	}

	err = u.service.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if err == service.ErrUserDuplicate {
		ctx.JSON(http.StatusOK, Result{
			Code: 3,
			Msg:  "重复邮箱，请换一个邮箱",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "注册成功",
	})
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.service.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "用户名或者密码不正确",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	sess := sessions.Default(ctx)
	sess.Set(userIdKey, user.Id)
	sess.Save()
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.service.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "用户名或者密码不正确",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if err = u.setJWTToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	fmt.Println(user)
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type Profile struct {
		Nickname  string `json:"nickname"`
		Birthday  string `json:"birthday"`
		Biography string `json:"biography"`
	}
	var req Profile
	if err := ctx.Bind(&req); err != nil {
		return
	}
	req.Nickname = strings.Trim(req.Nickname, " ")
	if len([]rune(req.Nickname)) > 20 {
		ctx.JSON(http.StatusOK, Result{
			Code: 1,
			Msg:  "昵称长度不超过20",
		})
	}

	req.Biography = strings.Trim(req.Biography, " ")
	if len([]rune(req.Biography)) > 500 {
		ctx.JSON(http.StatusOK, Result{
			Code: 1,
			Msg:  "个人简介长度不超过500",
		})
	}

	req.Birthday = strings.Trim(req.Birthday, " ")
	if len([]rune(req.Birthday)) > 0 {
		isBirthday, err := u.birthdayRegexExp.MatchString(req.Birthday)
		if err != nil {
			ctx.JSON(http.StatusOK, Result{
				Code: 1,
				Msg:  "生日格式不正确",
			})
			return
		}
		if !isBirthday {
			ctx.JSON(http.StatusOK, Result{
				Code: 1,
				Msg:  "生日格式必须是1992-01-01",
			})
			return
		}
		_, err = time.Parse("2006-1-02", req.Birthday)
		if err != nil {
			ctx.JSON(http.StatusOK, Result{
				Code: 1,
				Msg:  "非法时间",
			})
			return
		}
	}

	// sess := sessions.Default(ctx)
	// value := sess.Get(userIdKey)
	uid, exists := ctx.Get(userIdKey)
	if !exists {
		ctx.JSON(http.StatusOK, Result{
			Code: http.StatusUnauthorized,
			Msg:  "未登录",
		})
		return
	}
	userId := uid.(int64)

	err := u.service.Edit(ctx, domain.User{
		Id:        userId,
		Nickname:  req.Nickname,
		Birthday:  req.Birthday,
		Biography: req.Biography,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 3,
			Msg:  "更新失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "修改成功",
	})
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email     string
		Nickname  string
		Birthday  string
		Biography string
	}
	value, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	id, _ := value.(int64)
	user, err := u.service.Profile(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	res := Profile{
		Email:     user.Email,
		Nickname:  user.Nickname,
		Birthday:  user.Birthday,
		Biography: user.Biography,
	}
	ctx.JSON(http.StatusOK, Result{
		Data: res,
		Msg:  "success",
	})
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 考虑正则表达式
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		return
	}
	err := u.codeSvc.Send(ctx, loginSmsTemplate, loginBiz, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := u.codeSvc.Verify(ctx, loginBiz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码有误",
		})
		return
	}
	user, err := u.service.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	if err = u.setJWTToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "验证码校验通过",
	})
}

func (u *UserHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Uid: uid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("O0GWcsczOJHHM8Pu6l2JD9ftliO4Xfou"))
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid int64
}
