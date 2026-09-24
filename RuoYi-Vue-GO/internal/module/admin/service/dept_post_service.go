// 部门 + 岗位业务（对位 SysDeptServiceImpl + SysPostServiceImpl + 两个 Controller）。
package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	bizerr "ruoyi-vue-go/internal/common/errors"
	"ruoyi-vue-go/internal/module/admin/dao"
)

// DeptService 部门。
type DeptService struct{ dao *dao.DeptDao }

func NewDeptService(d *dao.DeptDao) *DeptService { return &DeptService{dao: d} }

// List 列表（对位 selectDeptList，返回扁平列表由前端树化；Java 返回扁平+children 字段空）。
func (s *DeptService) List(ctx context.Context, deptName, status string) ([]dao.DeptDO, error) {
	return s.dao.SelectDeptList(ctx, deptName, status)
}

// ListExclude 排除自身与子孙（对位 excludeChild）。
func (s *DeptService) ListExclude(ctx context.Context, deptID int64) ([]dao.DeptDO, error) {
	list, err := s.dao.SelectDeptList(ctx, "", "")
	if err != nil {
		return nil, err
	}
	return dao.ExcludeChild(list, deptID), nil
}

// GetByID 详情（对位 getInfo；checkDeptDataScope 在 admin-only 阶段为空操作）。
func (s *DeptService) GetByID(ctx context.Context, deptID int64) (*dao.DeptDO, error) {
	return s.dao.SelectDeptById(ctx, deptID)
}

// DeptInput 新增/修改参数。
type DeptInput struct {
	DeptID   int64
	ParentID int64
	DeptName string
	OrderNum int
	Leader   string
	Phone    string
	Email    string
	Status   string
}

// Add 新增（对位 add + insertDept：同父同名拒绝、父停用拒绝、ancestors 拼接）。
func (s *DeptService) Add(ctx context.Context, in *DeptInput, operator string) error {
	info, _ := s.dao.CheckDeptNameUnique(ctx, in.DeptName, in.ParentID)
	if info != nil {
		return errf("新增部门'%s'失败，部门名称已存在", in.DeptName)
	}
	parent, err := s.dao.SelectDeptById(ctx, in.ParentID)
	if err != nil {
		return errf("新增部门'%s'失败，上级部门不存在", in.DeptName)
	}
	if parent.Status != "0" {
		return errf("部门停用，不允许新增")
	}
	m := &dao.DeptDO{
		ParentID: in.ParentID, DeptName: in.DeptName,
		Ancestors: parent.Ancestors + "," + itoa64(in.ParentID),
		OrderNum:  in.OrderNum, Leader: in.Leader, Phone: in.Phone, Email: in.Email,
		Status: in.Status, CreateBy: operator,
	}
	return s.dao.InsertDept(ctx, m)
}

// Edit 修改（对位 edit + updateDept 全部分支）。
func (s *DeptService) Edit(ctx context.Context, in *DeptInput, operator string) error {
	info, _ := s.dao.CheckDeptNameUnique(ctx, in.DeptName, in.ParentID)
	if info != nil && info.DeptID != in.DeptID {
		return errf("修改部门'%s'失败，部门名称已存在", in.DeptName)
	}
	if in.ParentID == in.DeptID {
		return errf("修改部门'%s'失败，上级部门不能是自己", in.DeptName)
	}
	if in.Status == "1" {
		n, _ := s.dao.SelectNormalChildrenDeptById(ctx, in.DeptID)
		if n > 0 {
			return errf("该部门包含未停用的子部门！")
		}
	}
	// ancestors 重算 + 子孙同步（对位 updateDept + updateDeptChildren）
	newParent, err1 := s.dao.SelectDeptById(ctx, in.ParentID)
	oldDept, err2 := s.dao.SelectDeptById(ctx, in.DeptID)
	newAncestors := oldDept.Ancestors
	if err1 == nil && err2 == nil {
		newAncestors = newParent.Ancestors + "," + itoa64(in.ParentID)
		if newAncestors != oldDept.Ancestors {
			children, _ := s.dao.SelectChildrenDeptById(ctx, in.DeptID)
			for i := range children {
				// 对位 replace：旧 ancestors 前缀替换为新前缀
				children[i].Ancestors = strings.Replace(children[i].Ancestors, oldDept.Ancestors, newAncestors, 1)
			}
			_ = s.dao.UpdateDeptChildren(ctx, children)
		}
	}
	m := &dao.DeptDO{
		DeptID: in.DeptID, ParentID: in.ParentID, DeptName: in.DeptName,
		Ancestors: newAncestors, OrderNum: in.OrderNum, Leader: in.Leader,
		Phone: in.Phone, Email: in.Email, Status: in.Status, CreateBy: operator,
	}
	if err := s.dao.UpdateDept(ctx, m); err != nil {
		return err
	}
	// 启用状态时启用全部上级（对位 updateParentDeptStatusNormal）
	if in.Status == "0" && newAncestors != "" && newAncestors != "0" {
		var ids []int64
		for _, p := range strings.Split(newAncestors, ",") {
			if id, e := strconv.ParseInt(p, 10, 64); e == nil {
				ids = append(ids, id)
			}
		}
		return s.dao.UpdateParentDeptStatusNormal(ctx, ids)
	}
	return nil
}

// Remove 删除（对位 remove：子部门/有用户 warn；dataScope 校验 admin 阶段空）。
// 返回 (warnMsg, error)。
func (s *DeptService) Remove(ctx context.Context, deptID int64) (string, error) {
	has, _ := s.dao.HasChildByDeptId(ctx, deptID)
	if has {
		return "存在下级部门,不允许删除", nil
	}
	hasUser, _ := s.dao.CheckDeptExistUser(ctx, deptID)
	if hasUser {
		return "部门存在用户,不允许删除", nil
	}
	return "", s.dao.DeleteDeptById(ctx, deptID)
}

// PostService 岗位。
type PostService struct{ dao *dao.PostDao }

func NewPostService(d *dao.PostDao) *PostService { return &PostService{dao: d} }

// List 分页列表（分页由 handler 挂 page.Paginate）。
func (s *PostService) List(ctx context.Context, postCode, postName, status string) ([]dao.PostDO, error) {
	return s.dao.SelectPostList(ctx, postCode, postName, status)
}

// GetByID 详情。
func (s *PostService) GetByID(ctx context.Context, postID int64) (*dao.PostDO, error) {
	return s.dao.SelectPostById(ctx, postID)
}

// Optionselect 选择框（登录即可，无权限串）。
func (s *PostService) Optionselect(ctx context.Context) ([]dao.PostDO, error) {
	return s.dao.SelectPostAll(ctx)
}

// PostInput 岗位参数。
type PostInput struct {
	PostID   int64
	PostCode string
	PostName string
	PostSort int
	Status   string
	Remark   string
}

// Add 新增（文案逐字对位）。
func (s *PostService) Add(ctx context.Context, in *PostInput, operator string) error {
	if info, _ := s.dao.CheckPostNameUnique(ctx, in.PostName); info != nil {
		return errf("新增岗位'%s'失败，岗位名称已存在", in.PostName)
	}
	if info, _ := s.dao.CheckPostCodeUnique(ctx, in.PostCode); info != nil {
		return errf("新增岗位'%s'失败，岗位编码已存在", in.PostName)
	}
	m := &dao.PostDO{PostCode: in.PostCode, PostName: in.PostName, PostSort: in.PostSort, Status: in.Status, Remark: in.Remark, CreateBy: operator}
	return s.dao.InsertPost(ctx, m)
}

// Edit 修改。
func (s *PostService) Edit(ctx context.Context, in *PostInput, operator string) error {
	if info, _ := s.dao.CheckPostNameUnique(ctx, in.PostName); info != nil && info.PostID != in.PostID {
		return errf("修改岗位'%s'失败，岗位名称已存在", in.PostName)
	}
	if info, _ := s.dao.CheckPostCodeUnique(ctx, in.PostCode); info != nil && info.PostID != in.PostID {
		return errf("修改岗位'%s'失败，岗位编码已存在", in.PostName)
	}
	m := &dao.PostDO{PostID: in.PostID, PostCode: in.PostCode, PostName: in.PostName, PostSort: in.PostSort, Status: in.Status, Remark: in.Remark, CreateBy: operator}
	return s.dao.UpdatePost(ctx, m)
}

// Remove 批量删除（分配检查文案对位 %s已分配,不能删除）。
func (s *PostService) Remove(ctx context.Context, postIDs []int64) error {
	for _, id := range postIDs {
		post, err := s.dao.SelectPostById(ctx, id)
		if err != nil {
			continue
		}
		n, _ := s.dao.CountUserPostById(ctx, id)
		if n > 0 {
			return errf("%s已分配,不能删除", post.PostName)
		}
	}
	return s.dao.DeletePostByIds(ctx, postIDs)
}

// errf 业务失败（code 500，对位 error(msg)）；warnf 警告（code 601，对位 warn(msg)）。
func errf(format string, args ...any) error {
	return bizerr.New(fmt.Sprintf(format, args...))
}

func warnf(format string, args ...any) error {
	return bizerr.NewWithCode(601, fmt.Sprintf(format, args...))
}

func itoa64(v int64) string { return strconv.FormatInt(v, 10) }
