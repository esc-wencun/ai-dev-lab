package enums

import "testing"

// 落库值必须显式等于 Java 枚举 ordinal（BusinessType.java / OperatorType.java /
// BusinessStatus.java / LimitType.java / UserStatus.java），逐一对表。

func TestBusinessType(t *testing.T) {
	cases := []struct {
		got      BusinessType
		wantVal  int
		wantName string
	}{
		{BusinessTypeOther, 0, "OTHER"},
		{BusinessTypeInsert, 1, "INSERT"},
		{BusinessTypeUpdate, 2, "UPDATE"},
		{BusinessTypeDelete, 3, "DELETE"},
		{BusinessTypeGrant, 4, "GRANT"},
		{BusinessTypeExport, 5, "EXPORT"},
		{BusinessTypeImport, 6, "IMPORT"},
		{BusinessTypeForce, 7, "FORCE"},
		{BusinessTypeGencode, 8, "GENCODE"},
		{BusinessTypeClean, 9, "CLEAN"},
	}
	for _, c := range cases {
		if int(c.got) != c.wantVal {
			t.Errorf("BusinessType 值与 Java ordinal 不一致: %s = %d, want %d", c.got, int(c.got), c.wantVal)
		}
		if c.got.String() != c.wantName {
			t.Errorf("BusinessType.String() = %q, want %q", c.got.String(), c.wantName)
		}
	}
}

func TestOperatorType(t *testing.T) {
	cases := []struct {
		got      OperatorType
		wantVal  int
		wantName string
	}{
		{OperatorTypeOther, 0, "OTHER"},
		{OperatorTypeManage, 1, "MANAGE"},
		{OperatorTypeMobile, 2, "MOBILE"},
	}
	for _, c := range cases {
		if int(c.got) != c.wantVal {
			t.Errorf("OperatorType 值与 Java ordinal 不一致: %s = %d, want %d", c.got, int(c.got), c.wantVal)
		}
		if c.got.String() != c.wantName {
			t.Errorf("OperatorType.String() = %q, want %q", c.got.String(), c.wantName)
		}
	}
}

func TestBusinessStatus(t *testing.T) {
	if int(BusinessStatusSuccess) != 0 || BusinessStatusSuccess.String() != "SUCCESS" {
		t.Errorf("BusinessStatusSuccess 不符: %d %q", int(BusinessStatusSuccess), BusinessStatusSuccess.String())
	}
	if int(BusinessStatusFail) != 1 || BusinessStatusFail.String() != "FAIL" {
		t.Errorf("BusinessStatusFail 不符: %d %q", int(BusinessStatusFail), BusinessStatusFail.String())
	}
}

func TestLimitType(t *testing.T) {
	if int(LimitTypeDefault) != 0 || LimitTypeDefault.String() != "DEFAULT" {
		t.Errorf("LimitTypeDefault 不符: %d %q", int(LimitTypeDefault), LimitTypeDefault.String())
	}
	if int(LimitTypeIP) != 1 || LimitTypeIP.String() != "IP" {
		t.Errorf("LimitTypeIP 不符: %d %q", int(LimitTypeIP), LimitTypeIP.String())
	}
}

func TestUserStatus(t *testing.T) {
	cases := []struct {
		got      UserStatus
		wantCode string
		wantInfo string
	}{
		{UserStatusOK, "0", "正常"},
		{UserStatusDisable, "1", "停用"},
		{UserStatusDeleted, "2", "删除"},
	}
	for _, c := range cases {
		if string(c.got) != c.wantCode {
			t.Errorf("UserStatus code = %q, want %q", string(c.got), c.wantCode)
		}
		if c.got.Info() != c.wantInfo {
			t.Errorf("UserStatus(%q).Info() = %q, want %q", string(c.got), c.got.Info(), c.wantInfo)
		}
	}
}

func TestResolveHttpMethod(t *testing.T) {
	for _, m := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "TRACE"} {
		got, ok := ResolveHttpMethod(m)
		if !ok || got != HttpMethod(m) {
			t.Errorf("ResolveHttpMethod(%q) = %q, %v; want %q, true", m, got, ok, m)
		}
	}
	if _, ok := ResolveHttpMethod("FOO"); ok {
		t.Error("ResolveHttpMethod(FOO) 应返回 false")
	}
	if _, ok := ResolveHttpMethod("get"); ok {
		t.Error("ResolveHttpMethod 区分大小写，小写应返回 false")
	}
}
