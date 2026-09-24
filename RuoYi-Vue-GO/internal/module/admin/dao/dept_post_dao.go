// 部门 + 岗位数据访问（SQL 逐字对位 SysDeptMapper.xml / SysPostMapper.xml）。
package dao

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ruoyi-vue-go/pkg/types"
)

// DeptDao 部门。
type DeptDao struct{ db *gorm.DB }

func NewDeptDao(db *gorm.DB) *DeptDao { return &DeptDao{db: db} }

// deptColumns 对位 selectDeptVo。
const deptColumns = `select d.dept_id, d.parent_id, d.ancestors, d.dept_name, d.order_num, d.leader, d.phone, d.email, d.status, d.del_flag, d.create_by, d.create_time from sys_dept d`

// DeptDO 部门扫描模型（含 Java SysDept 的完整响应字段）。
type DeptDO struct {
	DeptID     int64          `gorm:"column:dept_id" json:"deptId"`
	ParentID   int64          `gorm:"column:parent_id" json:"parentId"`
	Ancestors  string         `gorm:"column:ancestors" json:"ancestors"`
	DeptName   string         `gorm:"column:dept_name" json:"deptName"`
	OrderNum   int            `gorm:"column:order_num" json:"orderNum"`
	Leader     string         `gorm:"column:leader" json:"leader"`
	Phone      string         `gorm:"column:phone" json:"phone"`
	Email      string         `gorm:"column:email" json:"email"`
	Status     string         `gorm:"column:status" json:"status"`
	DelFlag    string         `gorm:"column:del_flag" json:"delFlag"`
	CreateBy   string         `gorm:"column:create_by" json:"createBy"`
	CreateTime types.DateTime `gorm:"column:create_time" json:"createTime"`
	Children   []*DeptDO      `gorm:"-" json:"children,omitempty"`
}

// SelectDeptList 条件查询（对位 selectDeptList 动态 where + order by parent_id, order_num）。
func (d *DeptDao) SelectDeptList(ctx context.Context, deptName, status string) ([]DeptDO, error) {
	where := "where d.del_flag = '0'"
	var args []any
	if deptName != "" {
		where += " AND dept_name like concat('%', ?, '%')"
		args = append(args, deptName)
	}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}
	var list []DeptDO
	err := d.db.WithContext(ctx).Raw(deptColumns+" "+where+" order by d.parent_id, d.order_num", args...).Scan(&list).Error
	return list, err
}

// SelectDeptById 按 ID（对位 selectDeptById）。
func (d *DeptDao) SelectDeptById(ctx context.Context, deptID int64) (*DeptDO, error) {
	var m DeptDO
	err := d.db.WithContext(ctx).Raw(deptColumns+" where d.dept_id = ?", deptID).Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m.DeptID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &m, nil
}

// ExcludeChild 排除自身与子孙（对位 excludeChild 的 removeIf 逻辑，Java 在内存过滤，Go 同样）。
func ExcludeChild(list []DeptDO, deptID int64) []DeptDO {
	var ret []DeptDO
	for _, m := range list {
		if m.DeptID == deptID || strings.Contains(","+m.Ancestors+",", fmt.Sprintf(",%d,", deptID)) {
			continue
		}
		ret = append(ret, m)
	}
	return ret
}

// CheckDeptNameUnique 同父下同名（对位 checkDeptNameUnique limit 1）。
func (d *DeptDao) CheckDeptNameUnique(ctx context.Context, deptName string, parentID int64) (*DeptDO, error) {
	var m DeptDO
	err := d.db.WithContext(ctx).Raw(deptColumns+" where d.dept_name = ? and d.parent_id = ? and d.del_flag = '0' limit 1", deptName, parentID).Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m.DeptID == 0 {
		return nil, nil
	}
	return &m, nil
}

// HasChildByDeptId 有子节点（对位 hasChildByDeptId）。
func (d *DeptDao) HasChildByDeptId(ctx context.Context, deptID int64) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(1) from sys_dept where del_flag = '0' and parent_id = ? limit 1", deptID).Scan(&n).Error
	return n > 0, err
}

// CheckDeptExistUser 部门有用户（对位 checkDeptExistUser）。
func (d *DeptDao) CheckDeptExistUser(ctx context.Context, deptID int64) (bool, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(1) from sys_user where dept_id = ? and del_flag = '0'", deptID).Scan(&n).Error
	return n > 0, err
}

// SelectNormalChildrenDeptById 正常状态子孙数（对位同名）。
func (d *DeptDao) SelectNormalChildrenDeptById(ctx context.Context, deptID int64) (int64, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(*) from sys_dept where status = 0 and del_flag = '0' and find_in_set(?, ancestors)", deptID).Scan(&n).Error
	return n, err
}

// InsertDept 新增（对位 insertDept；ancestors 由 service 计算传入）。
func (d *DeptDao) InsertDept(ctx context.Context, m *DeptDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_dept (parent_id, dept_name, ancestors, order_num, leader, phone, email, status, del_flag, create_by, create_time) values (?, ?, ?, ?, ?, ?, ?, ?, '0', ?, now())",
		m.ParentID, m.DeptName, m.Ancestors, m.OrderNum, m.Leader, m.Phone, m.Email, m.Status, m.CreateBy).Error
}

// UpdateDept 修改（对位 updateDept 动态 set 的常用字段全集）。
func (d *DeptDao) UpdateDept(ctx context.Context, m *DeptDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_dept set parent_id = ?, ancestors = ?, dept_name = ?, order_num = ?, leader = ?, phone = ?, email = ?, status = ?, update_by = ?, update_time = now() where dept_id = ?",
		m.ParentID, m.Ancestors, m.DeptName, m.OrderNum, m.Leader, m.Phone, m.Email, m.Status, m.CreateBy, m.DeptID).Error
}

// UpdateDeptChildren 批量更新子孙 ancestors（对位 updateDeptChildren 的 case when）。
func (d *DeptDao) UpdateDeptChildren(ctx context.Context, children []DeptDO) error {
	for _, c := range children {
		if err := d.db.WithContext(ctx).Exec("update sys_dept set ancestors = ? where dept_id = ?", c.Ancestors, c.DeptID).Error; err != nil {
			return err
		}
	}
	return nil
}

// SelectChildrenDeptById 全部子孙（对位 selectChildrenDeptById find_in_set）。
func (d *DeptDao) SelectChildrenDeptById(ctx context.Context, deptID int64) ([]DeptDO, error) {
	var list []DeptDO
	err := d.db.WithContext(ctx).Raw("select dept_id, parent_id, ancestors from sys_dept where find_in_set(?, ancestors)", deptID).Scan(&list).Error
	return list, err
}

// UpdateParentDeptStatusNormal 启用祖先链（对位 updateDeptStatusNormal）。
func (d *DeptDao) UpdateParentDeptStatusNormal(ctx context.Context, deptIDs []int64) error {
	if len(deptIDs) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(deptIDs)), ",")
	args := make([]any, len(deptIDs))
	for i, id := range deptIDs {
		args[i] = id
	}
	return d.db.WithContext(ctx).Exec("update sys_dept set status = '0' where dept_id in ("+ph+")", args...).Error
}

// DeleteDeptById 软删（对位 deleteDeptById）。
func (d *DeptDao) DeleteDeptById(ctx context.Context, deptID int64) error {
	return d.db.WithContext(ctx).Exec("update sys_dept set del_flag = '2' where dept_id = ?", deptID).Error
}

// PostDao 岗位。
type PostDao struct{ db *gorm.DB }

func NewPostDao(db *gorm.DB) *PostDao { return &PostDao{db: db} }

// postColumns 对位 selectPostVo。
const postColumns = `select post_id, post_code, post_name, post_sort, status, create_by, create_time, remark from sys_post`

// PostDO 岗位模型。
type PostDO struct {
	PostID     int64          `gorm:"column:post_id" json:"postId"`
	PostCode   string         `gorm:"column:post_code" json:"postCode"`
	PostName   string         `gorm:"column:post_name" json:"postName"`
	PostSort   int            `gorm:"column:post_sort" json:"postSort"`
	Status     string         `gorm:"column:status" json:"status"`
	CreateBy   string         `gorm:"column:create_by" json:"createBy"`
	CreateTime types.DateTime `gorm:"column:create_time" json:"createTime"`
	Remark     string         `gorm:"column:remark" json:"remark"`
}

// SelectPostList 条件列表（对位 selectPostList：postCode/postName like、status；排序对位 order by post_sort）。
func (d *PostDao) SelectPostList(ctx context.Context, postCode, postName, status string) ([]PostDO, error) {
	where := "where 1=1"
	var args []any
	if postCode != "" {
		where += " AND post_code like concat('%', ?, '%')"
		args = append(args, postCode)
	}
	if postName != "" {
		where += " AND post_name like concat('%', ?, '%')"
		args = append(args, postName)
	}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}
	var list []PostDO
	err := d.db.WithContext(ctx).Raw(postColumns+" "+where+" order by post_sort", args...).Scan(&list).Error
	return list, err
}

// SelectPostAll 全部（optionselect 用）。
func (d *PostDao) SelectPostAll(ctx context.Context) ([]PostDO, error) {
	var list []PostDO
	err := d.db.WithContext(ctx).Raw(postColumns + " order by post_sort").Scan(&list).Error
	return list, err
}

// SelectPostById 按 ID。
func (d *PostDao) SelectPostById(ctx context.Context, postID int64) (*PostDO, error) {
	var m PostDO
	err := d.db.WithContext(ctx).Raw(postColumns+" where post_id = ?", postID).Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m.PostID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &m, nil
}

// CheckPostNameUnique / CheckPostCodeUnique 唯一性（对位同名 limit 1）。
func (d *PostDao) CheckPostNameUnique(ctx context.Context, name string) (*PostDO, error) {
	var m PostDO
	err := d.db.WithContext(ctx).Raw(postColumns+" where post_name = ? limit 1", name).Scan(&m).Error
	if err != nil || m.PostID == 0 {
		return nil, err
	}
	return &m, nil
}

func (d *PostDao) CheckPostCodeUnique(ctx context.Context, code string) (*PostDO, error) {
	var m PostDO
	err := d.db.WithContext(ctx).Raw(postColumns+" where post_code = ? limit 1", code).Scan(&m).Error
	if err != nil || m.PostID == 0 {
		return nil, err
	}
	return &m, nil
}

// CountUserPostById 岗位使用数（对位 countUserPostById）。
func (d *PostDao) CountUserPostById(ctx context.Context, postID int64) (int64, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(1) from sys_user_post where post_id = ?", postID).Scan(&n).Error
	return n, err
}

// InsertPost 新增。
func (d *PostDao) InsertPost(ctx context.Context, m *PostDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_post (post_code, post_name, post_sort, status, create_by, create_time, remark) values (?, ?, ?, ?, ?, now(), ?)",
		m.PostCode, m.PostName, m.PostSort, m.Status, m.CreateBy, m.Remark).Error
}

// UpdatePost 修改。
func (d *PostDao) UpdatePost(ctx context.Context, m *PostDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_post set post_code = ?, post_name = ?, post_sort = ?, status = ?, update_by = ?, update_time = now(), remark = ? where post_id = ?",
		m.PostCode, m.PostName, m.PostSort, m.Status, m.CreateBy, m.Remark, m.PostID).Error
}

// DeletePostByIds 批量软删（对位 deletePostByIds 物理 delete；Java 为物理删除）。
func (d *PostDao) DeletePostByIds(ctx context.Context, postIDs []int64) error {
	ph := strings.TrimSuffix(strings.Repeat("?,", len(postIDs)), ",")
	args := make([]any, len(postIDs))
	for i, id := range postIDs {
		args[i] = id
	}
	return d.db.WithContext(ctx).Exec("delete from sys_post where post_id in ("+ph+")", args...).Error
}

// guard
