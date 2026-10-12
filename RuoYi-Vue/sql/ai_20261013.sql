-- ============================================================
-- AI 对话正式化：菜单权限收敛（specs/2.0.3）
-- 可重复执行；本机执行：docker exec -i ruoyi-mysql mysql -uwencun -p111111 ry-vue < sql/ai_20261013.sql
-- ============================================================

SET NAMES utf8mb4;

-- 1. 菜单 2020「AI对话」补权限点（进入对话页即授权对话功能）
update sys_menu set perms = 'ai:chat:chat' where menu_id = 2020;

-- 2. 智能体管理菜单（2030，hidden 占位——管理页 UI 列 2.0.4，先落权限点防 500）
delete from sys_menu where menu_id between 2030 and 2039;
insert into sys_menu values('2030', '智能体管理', '2000', '5', 'agent', 'ai/agent/index', '', '', 1, 0, 'C', '1', '0', 'ai:agent:list', 'education', 'admin', sysdate(), '', null, 'AI 智能体管理（管理页 2.0.4）');
insert into sys_menu values('2031', '智能体查询', '2030', '1', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:agent:query',  '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2032', '智能体新增', '2030', '2', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:agent:add',    '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2033', '智能体修改', '2030', '3', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:agent:edit',   '#', 'admin', sysdate(), '', null, '');
insert into sys_menu values('2034', '智能体删除', '2030', '4', '', '', '', '', 1, 0, 'F', '0', '0', 'ai:agent:remove', '#', 'admin', sysdate(), '', null, '');
