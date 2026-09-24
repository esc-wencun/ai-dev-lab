// Package page 分页与排序（对位 Java PageHelper startPage + PageDomain + SqlUtil + TableDataInfo）。
//
// 契约（与 Java/Python 版一致）：
//   - query params：pageNum（默认 1）、pageSize（默认 10）、orderByColumn（驼峰）、isAsc（asc/desc）；
//   - orderByColumn 驼峰转下划线后与 isAsc 拼成 order by 片段；
//   - 排序字段白名单校验（仅字母/数字/下划线/空格/逗号/点，长度 ≤500），拒绝注入；
//   - 分页响应顶层字段 rows/total/code/msg（对位 TableDataInfo，无 data 字段）。
package page

import (
	"context"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	bizerrors "ruoyi-vue-go/internal/common/errors"
)

// sqlPattern 排序字段白名单：仅支持字母、数字、下划线、空格、逗号、小数点（对位 SqlUtil.SQL_PATTERN）
const sqlPattern = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_ ,."

// orderByMaxLength 限制 orderBy 最大长度（对位 SqlUtil.ORDER_BY_MAX_LENGTH）
const orderByMaxLength = 500

// Domain 分页参数（对位 Java PageDomain）
type Domain struct {
	PageNum       int
	PageSize      int
	OrderByColumn string
	IsAsc         string
}

// FromGin 从 query params 读取分页参数（对位 TableSupport.buildPageRequest）
func FromGin(c *gin.Context) Domain {
	d := Domain{
		PageNum:       queryInt(c, "pageNum", 1),
		PageSize:      queryInt(c, "pageSize", 10),
		OrderByColumn: c.Query("orderByColumn"),
		IsAsc:         "asc",
	}
	if v := c.Query("isAsc"); v != "" {
		d.IsAsc = v
	}
	return d
}

// ToOrderBy 生成 order by 片段，如 "user_id desc"（对位 PageDomain.getOrderBy + SqlUtil.escapeOrderBySql）。
// 非法字段或超长返回 BusinessError（对位 UtilException）。
func (d Domain) ToOrderBy() (string, error) {
	column := camelToUnderScore(d.OrderByColumn)
	isAsc, err := normalizeAsc(d.IsAsc)
	if err != nil {
		return "", err
	}
	orderBy := strings.TrimSpace(column + " " + isAsc)
	if err := escapeOrderBySql(orderBy); err != nil {
		return "", err
	}
	return orderBy, nil
}

// Paginate 分页执行器（对位 PageHelper.startPage + BaseController.getDataTable 两步）。
//
// query：已含 Model/Where 的 gorm 会话；dest：&[]T{}；isPage=false 时查全表、total 为行数
// （对位 Java 未调 startPage 的场景）。Count 与 Find 各自从 query 克隆会话，互不污染。
func Paginate(ctx context.Context, query *gorm.DB, dest any, d Domain, isPage bool) (int64, error) {
	q := query.WithContext(ctx)

	if !isPage {
		if err := q.Session(&gorm.Session{}).Find(dest).Error; err != nil {
			return 0, err
		}
		return int64(reflect.ValueOf(dest).Elem().Len()), nil
	}

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, err
	}

	pageNum, pageSize := d.PageNum, d.PageSize
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	rowsQ := q.Session(&gorm.Session{})
	if d.OrderByColumn != "" {
		orderBy, err := d.ToOrderBy()
		if err != nil {
			return 0, err
		}
		rowsQ = rowsQ.Order(orderBy)
	}
	if err := rowsQ.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(dest).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// escapeOrderBySql 排序字段白名单校验（对位 SqlUtil.escapeOrderBySql）
func escapeOrderBySql(value string) error {
	if value == "" {
		return nil
	}
	for _, r := range value {
		if !strings.ContainsRune(sqlPattern, r) {
			return bizerrors.New("参数不符合规范，不能进行查询")
		}
	}
	if len(value) > orderByMaxLength {
		return bizerrors.New("参数已超过最大限制，不能进行查询")
	}
	return nil
}

// normalizeAsc 归一排序方向（对位 PageDomain.setIsAsc 的前端兼容：ascending/descending）。
// 比对 Java 更严格：未知方向直接拒绝（Java 会透传给 SQL），防注入。
func normalizeAsc(v string) (string, error) {
	switch v {
	case "", "asc", "ascending":
		return "asc", nil
	case "desc", "descending":
		return "desc", nil
	default:
		return "", bizerrors.New("参数不符合规范，不能进行查询")
	}
}

// camelToUnderScore 驼峰转下划线（对位 StringUtils.toUnderScoreCase：大写字母前插下划线再整体小写）
func camelToUnderScore(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// queryInt 读整型 query 参数，非法或缺省用默认值（对位 Convert.toInt(param, default)）
func queryInt(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n := 0
	for _, r := range v {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
		if n > 1<<30 {
			return def
		}
	}
	if n <= 0 {
		return def
	}
	return n
}
