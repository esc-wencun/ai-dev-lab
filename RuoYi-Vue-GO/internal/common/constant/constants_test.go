package constant

import "testing"

// 以下常量值均逐条对照 Java 源码（RuoYi-Vue/ruoyi-common/.../constant/*.java）抄录，
// 改动任何一处预期值前必须先核对 Java 版。

func TestCacheConstants(t *testing.T) {
	cases := []struct{ got, want string }{
		{LoginTokenKey, "login_tokens:"},
		{CaptchaCodeKey, "captcha_codes:"},
		{SysConfigKey, "sys_config:"},
		{SysDictKey, "sys_dict:"},
		{RepeatSubmitKey, "repeat_submit:"},
		{RateLimitKey, "rate_limit:"},
		{PwdErrCntKey, "pwd_err_cnt:"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("CacheConstants 与 Java 不一致: got %q, want %q", c.got, c.want)
		}
	}
}

func TestConstants(t *testing.T) {
	strCases := []struct{ got, want string }{
		{Success, "0"},
		{Fail, "1"},
		{LoginSuccess, "Success"},
		{Logout, "Logout"},
		{Register, "Register"},
		{LoginFail, "Error"},
		{AllPermission, "*:*:*"},
		{SuperAdmin, "admin"},
		{RoleDelimiter, ","},
		{PermissionDelimiter, ","},
		{Token, "token"},
		{TokenPrefix, "Bearer "},
		{LoginUserKey, "login_user_key"},
		{JwtUserid, "userid"},
		{JwtUsername, "sub"}, // io.jsonwebtoken Claims.SUBJECT
		{JwtAvatar, "avatar"},
		{JwtCreated, "created"},
		{JwtAuthorities, "authorities"},
		{ResourcePrefix, "/profile"},
		// Constants.Dept 数据权限范围
		{DataScopeAll, "1"},
		{DataScopeCustom, "2"},
		{DataScopeDept, "3"},
		{DataScopeDeptAndChild, "4"},
		{DataScopeSelf, "5"},
	}
	for _, c := range strCases {
		if c.got != c.want {
			t.Errorf("Constants 与 Java 不一致: got %q, want %q", c.got, c.want)
		}
	}

	if CaptchaExpiration != 2 {
		t.Errorf("CaptchaExpiration = %d, want 2", CaptchaExpiration)
	}
}

func TestUserConstants(t *testing.T) {
	strCases := []struct{ got, want string }{
		{SysUser, "SYS_USER"},
		{Normal, "0"},
		{Exception, "1"},
		{UserDisable, "1"},
		{RoleNormal, "0"},
		{RoleDisable, "1"},
		{DeptNormal, "0"},
		{DeptDisable, "1"},
		{DictNormal, "0"},
		{Yes, "Y"},
		{YesFrame, "0"},
		{NoFrame, "1"},
		{TypeDir, "M"},
		{TypeMenu, "C"},
		{TypeButton, "F"},
		{Layout, "Layout"},
		{ParentView, "ParentView"},
		{InnerLink, "InnerLink"},
	}
	for _, c := range strCases {
		if c.got != c.want {
			t.Errorf("UserConstants 与 Java 不一致: got %q, want %q", c.got, c.want)
		}
	}

	boolCases := []struct {
		name      string
		got, want bool
	}{
		{"Unique", Unique, true},
		{"NotUnique", NotUnique, false},
	}
	for _, c := range boolCases {
		if c.got != c.want {
			t.Errorf("UserConstants.%s = %v, want %v", c.name, c.got, c.want)
		}
	}

	intCases := []struct {
		name      string
		got, want int
	}{
		{"UsernameMinLength", UsernameMinLength, 2},
		{"UsernameMaxLength", UsernameMaxLength, 20},
		{"PasswordMinLength", PasswordMinLength, 5},
		{"PasswordMaxLength", PasswordMaxLength, 20},
	}
	for _, c := range intCases {
		if c.got != c.want {
			t.Errorf("UserConstants.%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}
