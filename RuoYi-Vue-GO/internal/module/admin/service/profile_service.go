// 个人中心 / 注册 业务（对位 SysProfileController + SysRegisterService 的业务逻辑）。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/common/message"
	"ruoyi-vue-go/internal/common/model"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/security"
	"ruoyi-vue-go/pkg/types"
)

// ProfileService 个人中心。
type ProfileService struct {
	profileDao *dao.ProfileDao
	loginDao   *dao.LoginDao
	cache      *cache.RedisCache
	token      *security.TokenService
	login      *LoginService // 复用验证码校验与参数配置查询
}

// NewProfileService 构造。
func NewProfileService(p *dao.ProfileDao, l *dao.LoginDao, rc *cache.RedisCache, ls *LoginService) *ProfileService {
	return &ProfileService{profileDao: p, loginDao: l, cache: rc, login: ls}
}

// SetToken 注入 TokenService（会话刷新用，规避构造环）。
func (s *ProfileService) SetToken(t *security.TokenService) { s.token = t }

// ProfileResult GET profile 响应数据（user 原始 JSON + roleGroup/postGroup）。
type ProfileResult struct {
	UserJSON  json.RawMessage
	RoleGroup string
	PostGroup string
}

// GetProfile 查询个人信息（对位 profile()：会话 user + roleGroup/postGroup 逗号分隔）。
func (s *ProfileService) GetProfile(ctx context.Context, lu *model.LoginUser) (*ProfileResult, error) {
	roleNames, _ := s.profileDao.RoleNamesByUser(ctx, lu.Username())
	postNames, _ := s.profileDao.PostNamesByUser(ctx, lu.Username())
	return &ProfileResult{
		UserJSON:  lu.User,
		RoleGroup: strings.Join(roleNames, ","),
		PostGroup: strings.Join(postNames, ","),
	}, nil
}

// UpdateProfileInput 修改个人信息（对位 updateProfile 仅四字段）。
type UpdateProfileInput struct {
	NickName    string
	Email       string
	Phonenumber string
	Sex         string
}

// UpdateProfile 修改个人信息（唯一性校验文案逐字对位 Java）。
func (s *ProfileService) UpdateProfile(ctx context.Context, lu *model.LoginUser, in *UpdateProfileInput) error {
	if in.Phonenumber != "" {
		ok, _ := s.profileDao.CheckPhoneUnique(ctx, in.Phonenumber, lu.UserId)
		if !ok {
			return fmt.Errorf("修改用户'%s'失败，手机号码已存在", lu.Username())
		}
	}
	if in.Email != "" {
		ok, _ := s.profileDao.CheckEmailUnique(ctx, in.Email, lu.UserId)
		if !ok {
			return fmt.Errorf("修改用户'%s'失败，邮箱账号已存在", lu.Username())
		}
	}
	return s.profileDao.UpdateProfile(ctx, lu.UserId, in.NickName, in.Email, in.Phonenumber, in.Sex)
}

// UpdatePwd 修改密码（对位 updatePwd：旧密码错误/新旧相同文案；成功更新 pwd_update_date）。
func (s *ProfileService) UpdatePwd(ctx context.Context, lu *model.LoginUser, oldPassword, newPassword string) error {
	ud, err := s.loginDao.LoadUserByName(ctx, lu.Username())
	if err != nil {
		return fmt.Errorf("修改密码异常，请联系管理员")
	}
	if bcrypt.CompareHashAndPassword([]byte(ud.User.Password), []byte(oldPassword)) != nil {
		return fmt.Errorf("修改密码失败，旧密码错误")
	}
	if bcrypt.CompareHashAndPassword([]byte(ud.User.Password), []byte(newPassword)) == nil {
		return fmt.Errorf("新密码不能与旧密码相同")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return fmt.Errorf("修改密码异常，请联系管理员")
	}
	if err := s.profileDao.ResetUserPwd(ctx, lu.UserId, string(hash)); err != nil {
		return fmt.Errorf("修改密码异常，请联系管理员")
	}
	return nil
}

// RefreshSessionUser 重新加载用户并回写会话（对位 tokenService.setLoginUser(loginUser)）。
func (s *ProfileService) RefreshSessionUser(ctx context.Context, lu *model.LoginUser) error {
	ud, err := s.loginDao.LoadUserByName(ctx, lu.Username())
	if err != nil {
		return err
	}
	view := SessionUserView(ud)
	data, err := json.Marshal(view)
	if err != nil {
		return err
	}
	lu.User = data
	return s.token.RefreshToken(ctx, lu)
}

// UpdateAvatar 更新头像 URL 并刷新会话（对位 avatar 的落库 + setLoginUser）。
func (s *ProfileService) UpdateAvatar(ctx context.Context, lu *model.LoginUser, avatarURL string) error {
	if err := s.profileDao.UpdateAvatar(ctx, lu.UserId, avatarURL); err != nil {
		return err
	}
	return s.RefreshSessionUser(ctx, lu)
}

// RegisterInput 注册参数（对位 RegisterBody）。
type RegisterInput struct {
	Username string
	Password string
	Code     string
	UUID     string
	IP       string
}

// Register 注册（对位 SysRegisterService.register：开关→验证码→校验链→落库；返回错误文案，空串=成功）。
func (s *ProfileService) Register(ctx context.Context, in *RegisterInput) string {
	// 开关：sys.account.registerUser == "true"
	v, _ := s.login.ConfigValue(ctx, "sys.account.registerUser")
	if v != "true" {
		return "当前系统没有开启注册功能！"
	}
	// 验证码（开关开启时；对位 SysRegisterService.validateCaptcha：取到即删、忽略大小写）
	if s.login.CaptchaEnabled(ctx) {
		key := cache.BuildKey(constant.CaptchaCodeKey, in.UUID)
		var stored string
		if err := s.cache.GetObject(ctx, key, &stored); err != nil {
			return message.UserJcaptchaExpire
		}
		_, _ = s.cache.Delete(ctx, key)
		if !strings.EqualFold(in.Code, stored) {
			return message.UserJcaptchaError
		}
	}
	// 校验链（文案逐字对位）
	if in.Username == "" {
		return "用户名不能为空"
	}
	if in.Password == "" {
		return "用户密码不能为空"
	}
	if l := len(in.Username); l < 2 || l > 20 {
		return "账户长度必须在2到20个字符之间"
	}
	if l := len(in.Password); l < 5 || l > 20 {
		return "密码长度必须在5到20个字符之间"
	}
	ok, _ := s.profileDao.CheckUserNameUnique(ctx, in.Username)
	if !ok {
		return fmt.Sprintf("保存用户'%s'失败，注册账号已存在", in.Username)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return "注册失败,请联系系统管理人员"
	}
	if err := s.profileDao.InsertUser(ctx, in.Username, in.Username, string(hash), types.Now()); err != nil {
		return "注册失败,请联系系统管理人员"
	}
	// 注册成功日志（对位 recordLogininfor(username, REGISTER, 注册成功)）
	s.login.RecordLogininforUA(ctx, in.Username, constant.Register, message.UserRegisterSuccess, in.IP, "", "")
	return "" // 成功
}
