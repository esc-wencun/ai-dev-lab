// 菜单路由构建（对位 SysMenuServiceImpl.selectMenuTreeByUserId + buildMenus + RouterVo/MetaVo）。
package service

import (
	"context"
	"strings"

	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/model/do"
)

// MenuService 菜单路由。
type MenuService struct {
	dao *dao.LoginDao
}

func NewMenuService(d *dao.LoginDao) *MenuService { return &MenuService{dao: d} }

// RouterVo 前端路由（对位 RouterVo；@JsonInclude(NON_EMPTY) 语义=空值字段不输出，
// Go 侧用 omitempty + 指针/字符串空值实现同等形态）。
type RouterVo struct {
	Name       string      `json:"name,omitempty"`
	Path       string      `json:"path"`
	Hidden     bool        `json:"hidden"`
	Redirect   string      `json:"redirect,omitempty"`
	Component  string      `json:"component,omitempty"`
	Query      string      `json:"query,omitempty"`
	AlwaysShow *bool       `json:"alwaysShow,omitempty"`
	Meta       *MetaVo     `json:"meta,omitempty"`
	Children   []*RouterVo `json:"children,omitempty"`
}

// MetaVo 路由显示信息（对位 MetaVo；link 仅 http(s):// 开头时输出）。
type MetaVo struct {
	Title   string  `json:"title"`
	Icon    string  `json:"icon"`
	NoCache bool    `json:"noCache"`
	Link    *string `json:"link,omitempty"` // NON_NULL：nil 不输出
}

// menuRootId 顶级菜单父 ID（对位 MENU_ROOT_ID）。
const menuRootId = 0

// GetMenuTreeByUser 用户菜单树（admin 全量，非 admin 按角色；对位 selectMenuTreeByUserId + getChildPerms）。
func (s *MenuService) GetMenuTreeByUser(ctx context.Context, userID int64) ([]*do.SysMenu, error) {
	var menus []do.SysMenu
	var err error
	if userID == AdminUserID {
		menus, err = s.dao.MenuTreeAll(ctx)
	} else {
		menus, err = s.dao.MenuTreeByUser(ctx, userID)
	}
	if err != nil {
		return nil, err
	}
	list := make([]*do.SysMenu, len(menus))
	for i := range menus {
		list[i] = &menus[i]
	}
	return buildMenuTree(list, menuRootId), nil
}

// buildMenuTree 菜单树构建（对位 getChildPerms：parentId 挂 children，保持 order by 顺序）。
func buildMenuTree(menus []*do.SysMenu, parentID int64) []*do.SysMenu {
	var ret []*do.SysMenu
	for _, m := range menus {
		if m.ParentID == parentID {
			m.Children = buildMenuTree(menus, m.MenuID)
			ret = append(ret, m)
		}
	}
	return ret
}

// BuildRouters 构建前端路由（对位 buildMenus，三分支结构逐条对齐）。
func (s *MenuService) BuildRouters(menus []*do.SysMenu) []*RouterVo {
	var routers []*RouterVo
	for _, menu := range menus {
		router := &RouterVo{
			Hidden:    menu.Visible == "1",
			Name:      routeName(menu.RouteName, menu.Path),
			Path:      routerPath(menu),
			Component: component(menu),
			Query:     menu.Query,
			Meta:      newMeta(menu.MenuName, menu.Icon, menu.IsCache == "1", menu.Path),
		}
		if len(menu.Children) > 0 && menu.MenuType == constantTypeDir {
			alwaysShow := true
			router.AlwaysShow = &alwaysShow
			router.Redirect = "noRedirect"
			router.Children = s.BuildRouters(menu.Children)
		} else if isMenuFrame(menu) {
			router.Meta = nil // 对位 router.setMeta(null)
			child := &RouterVo{
				Path:      menu.Path,
				Component: menu.Component,
				Name:      routeName(menu.RouteName, menu.Path),
				Meta:      newMeta(menu.MenuName, menu.Icon, menu.IsCache == "1", menu.Path),
				Query:     menu.Query,
			}
			router.Children = []*RouterVo{child}
		} else if menu.ParentID == menuRootId && isInnerLink(menu) {
			meta := newMeta(menu.MenuName, menu.Icon, false, "")
			router.Meta = meta
			router.Path = "/"
			routerPathClean := innerLinkReplaceEach(menu.Path)
			child := &RouterVo{
				Path:      routerPathClean,
				Component: constantInnerLink,
				Name:      routeName(menu.RouteName, routerPathClean),
				Meta:      newMetaLink(menu.MenuName, menu.Icon, menu.Path),
			}
			router.Children = []*RouterVo{child}
		}
		routers = append(routers, router)
	}
	return routers
}

// routeName 路由名（对位 getRouteName：优先 routeName，否则 capitalize(path)）。
func routeName(name, path string) string {
	if name != "" {
		return capitalize(name)
	}
	return capitalize(path)
}

// capitalize 首字母大写（对位 StringUtils.capitalize）。
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// routerPath 路由地址（对位 getRouterPath）。
func routerPath(menu *do.SysMenu) string {
	routerPath := menu.Path
	if menu.ParentID != menuRootId && isInnerLink(menu) {
		routerPath = innerLinkReplaceEach(routerPath)
	}
	if menu.ParentID == menuRootId && menu.MenuType == constantTypeDir && menu.IsFrame == constantNoFrame {
		routerPath = "/" + menu.Path
	} else if isMenuFrame(menu) {
		routerPath = "/"
	}
	return routerPath
}

// component 组件（对位 getComponent：Layout / 菜单自带 / InnerLink / ParentView）。
func component(menu *do.SysMenu) string {
	component := constantLayout
	if menu.Component != "" && !isMenuFrame(menu) {
		component = menu.Component
	} else if menu.Component == "" && menu.ParentID != menuRootId && isInnerLink(menu) {
		component = constantInnerLink
	} else if menu.Component == "" && isParentView(menu) {
		component = constantParentView
	}
	return component
}

// isMenuFrame 非外链一级菜单（对位 isMenuFrame）。
func isMenuFrame(menu *do.SysMenu) bool {
	return menu.ParentID == menuRootId && menu.MenuType == constantTypeMenu && menu.IsFrame == constantNoFrame
}

// isInnerLink 内链（对位 isInnerLink：非外链且 path 为 http(s)://）。
func isInnerLink(menu *do.SysMenu) bool {
	return menu.IsFrame == constantNoFrame && isHTTP(menu.Path)
}

// isParentView 非一级目录（对位 isParentView）。
func isParentView(menu *do.SysMenu) bool {
	return menu.ParentID != menuRootId && menu.MenuType == constantTypeDir
}

// isHTTP http(s):// 开头（对位 StringUtils.ishttp）。
func isHTTP(link string) bool {
	return strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://")
}

// innerLinkReplaceEach 内链地址转路由（对位同名：去协议/www，. 与 : 转 /）。
func innerLinkReplaceEach(path string) string {
	r := strings.NewReplacer("http://", "", "https://", "", "www.", "", ".", "/", ":", "/")
	return r.Replace(path)
}

// newMeta 常规 meta（对位 MetaVo(title, icon, noCache, link)：link 仅 http(s) 时输出）。
func newMeta(title, icon string, noCache bool, link string) *MetaVo {
	m := &MetaVo{Title: title, Icon: icon, NoCache: noCache}
	if isHTTP(link) {
		m.Link = &link
	}
	return m
}

// newMetaLink 强制 link 的 meta（对位 MetaVo(title, icon, path) 三参构造，内链子路由用）。
func newMetaLink(title, icon, link string) *MetaVo {
	m := &MetaVo{Title: title, Icon: icon, NoCache: false}
	if isHTTP(link) {
		m.Link = &link
	}
	return m
}

// 常量对位（复用 internal/common/constant 会与常量名冲突，就地短别名）。
const (
	constantTypeDir    = "M"
	constantTypeMenu   = "C"
	constantNoFrame    = "1"
	constantLayout     = "Layout"
	constantParentView = "ParentView"
	constantInnerLink  = "InnerLink"
)
