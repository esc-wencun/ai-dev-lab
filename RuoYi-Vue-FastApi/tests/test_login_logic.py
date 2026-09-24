"""
登录模块纯逻辑单元测试（不依赖数据库和Redis）
运行：python -m pytest test_login_logic.py -v
"""
import sys
import os
import base64
import hashlib
import hmac
from datetime import datetime

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from module_admin.entity.do.entity import SysMenu
from module_admin.service.login_service import build_menus, get_child_perms
from utils.pwd_util import PwdUtil


def make_menu(**kwargs):
    menu = SysMenu(
        menu_id=kwargs.get('menu_id', 1),
        menu_name=kwargs.get('menu_name', '菜单'),
        parent_id=kwargs.get('parent_id', 0),
        order_num=kwargs.get('order_num', 1),
        path=kwargs.get('path', 'test'),
        component=kwargs.get('component'),
        query=kwargs.get('query'),
        route_name=kwargs.get('route_name', ''),
        is_frame=kwargs.get('is_frame', 1),
        is_cache=kwargs.get('is_cache', 0),
        menu_type=kwargs.get('menu_type', 'C'),
        visible=kwargs.get('visible', '0'),
        status=kwargs.get('status', '0'),
        perms=kwargs.get('perms'),
        icon=kwargs.get('icon', '#'),
    )
    return menu


class TestPasswordCompat:
    """验证与java版BCryptPasswordEncoder的兼容性"""

    def test_verify_java_bcrypt_hash(self):
        # java版sql脚本中admin用户的密码哈希（明文admin123）
        java_hash = '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2'
        assert PwdUtil.verify_password('admin123', java_hash) is True

    def test_wrong_password(self):
        java_hash = '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2'
        assert PwdUtil.verify_password('wrong', java_hash) is False

    def test_hash_and_verify_roundtrip(self):
        hashed = PwdUtil.get_password_hash('test123')
        assert PwdUtil.verify_password('test123', hashed) is True
        assert hashed.startswith('$2b$') or hashed.startswith('$2a$')


class TestBuildMenus:
    """验证路由构建与java版buildMenus行为一致"""

    def test_admin_full_tree(self):
        # 一级目录
        system_dir = make_menu(menu_id=1, menu_name='系统管理', parent_id=0, order_num=1,
                               path='system', menu_type='M', icon='system')
        # 二级菜单
        user_menu = make_menu(menu_id=100, menu_name='用户管理', parent_id=1, order_num=1,
                              path='user', component='system/user/index', menu_type='C', perms='system:user:list')
        menus = [system_dir, user_menu]
        routers = build_menus(menus)
        assert len(routers) == 1
        router = routers[0]
        assert router['name'] == 'System'
        assert router['path'] == '/system'
        assert router['component'] == 'Layout'
        assert router['alwaysShow'] is True
        assert router['redirect'] == 'noRedirect'
        assert router['meta']['title'] == '系统管理'
        children = router['children']
        assert len(children) == 1
        child = children[0]
        assert child['path'] == 'user'
        assert child['component'] == 'system/user/index'
        assert child['name'] == 'User'
        assert child['meta']['title'] == '用户管理'

    def test_root_menu_frame(self):
        # 一级菜单（parent_id=0，类型C，非外链）应生成 Layout 嵌套结构
        menu = make_menu(menu_id=5, menu_name='首页', parent_id=0, path='index',
                         component='index', menu_type='C', is_frame=1)
        routers = build_menus([menu])
        router = routers[0]
        assert router['path'] == '/'
        assert router['component'] == 'Layout'
        assert router['meta'] is None
        assert len(router['children']) == 1
        assert router['children'][0]['path'] == 'index'
        assert router['children'][0]['component'] == 'index'

    def test_outer_link_root(self):
        # 顶级外链目录：is_frame=0（是外链）
        link = make_menu(menu_id=4, menu_name='若依官网', parent_id=0, path='http://ruoyi.vip',
                         menu_type='M', is_frame=0, icon='guide')
        routers = build_menus([link])
        router = routers[0]
        # 外链目录：is_frame=0不满足isMenuFrame/isInnerLink（两者都要求NO_FRAME=1），
        # 是一个普通路由，meta中带link，无children
        assert router['path'] == 'http://ruoyi.vip'
        assert router['hidden'] is False
        assert router['meta']['link'] == 'http://ruoyi.vip'
        assert 'children' not in router or not router.get('children')

    def test_inner_link(self):
        # 二级内链（http开头，is_frame=1表示非外链按钮->走InnerLink组件）
        menu = make_menu(menu_id=6, menu_name='监控页', parent_id=1, path='http://monitor.example.com',
                         menu_type='C', is_frame=1)
        parent = make_menu(menu_id=1, menu_name='目录', parent_id=0, path='dir', menu_type='M')
        # 与真实调用一致：传入数据库返回的平面列表
        routers = build_menus([parent, menu])
        children = routers[0].get('children') or []
        assert len(children) == 1
        assert children[0]['component'] == 'InnerLink'
        assert children[0]['meta']['link'] == 'http://monitor.example.com'

    def test_hidden_menu(self):
        menu = make_menu(menu_id=7, visible='1')
        routers = build_menus([menu])
        assert routers[0]['hidden'] is True

    def test_route_name_capitalize(self):
        # 非根级菜单才使用route_name/path驼峰化（与java版getRouteName一致）
        menu = make_menu(menu_id=8, menu_name='大分类', parent_id=1, route_name='', path='bigType',
                         menu_type='C')
        parent = make_menu(menu_id=1, menu_name='目录', parent_id=0, path='dir', menu_type='M')
        routers = build_menus([parent, menu])
        child = routers[0]['children'][0]
        assert child['name'] == 'BigType'
        # 配置了route_name时优先使用
        menu2 = make_menu(menu_id=9, menu_name='其他', parent_id=2, route_name='Custom', path='other',
                          menu_type='C')
        parent2 = make_menu(menu_id=2, menu_name='目录2', parent_id=0, path='dir2', menu_type='M')
        routers2 = build_menus([parent2, menu2])
        assert routers2[0]['children'][0]['name'] == 'Custom'

    def test_root_menu_frame_name_empty(self):
        # 根级菜单(isMenuFrame)路由name为空串（与java版getRouteName一致）
        menu = make_menu(menu_id=10, menu_name='工作台', parent_id=0, path='dashboard',
                         component='dashboard', menu_type='C', is_frame=1)
        routers = build_menus([menu])
        assert routers[0]['name'] == ''
        assert routers[0]['children'][0]['name'] == 'Dashboard'


class TestTreeBuilding:
    """验证菜单树构建"""

    def test_get_child_perms(self):
        root = make_menu(menu_id=1, parent_id=0)
        child1 = make_menu(menu_id=2, parent_id=1)
        child2 = make_menu(menu_id=3, parent_id=1)
        grandchild = make_menu(menu_id=4, parent_id=3)
        menus = [root, child1, child2, grandchild]
        tree = get_child_perms(menus, 0)
        assert len(tree) == 1
        assert len(tree[0].children) == 2
        assert len(tree[0].children[1].children) == 1


if __name__ == '__main__':
    import pytest
    sys.exit(pytest.main([__file__, '-v']))
