# Checklist · 2.0.4 智能体管理页

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [ ] suggestedQuestions 出口修复：options/list/getInfo 三口返回数组（curl 实测），对话页 chips 正常渲染（真实接口路径）
- [ ] 管理页 CRUD 全通：新增/编辑/删除/搜索/名称查重 500/被引用拒删（报错带数量）
- [ ] 建议问题动态行编辑：数组编辑 → 存库 JSON 字符串 → 回显数组 → 对话页 chips 消费
- [ ] 菜单 2030 放出：侧边栏「AI管理 → 智能体管理」可见可用；SQL 幂等
- [ ] 私有智能体可见性：options 下拉仅公开 + 本人私有
- [ ] 回归零影响：对话页（对话/流式/参数弹窗）+ 2.0.1 三管理页
- [ ] 测试数据清理；specs/README 状态收口
- [ ] javac 全模块 + 单测全绿（含新转换工具的容错单测）

## 遗留与决策记录

- 挂载编辑器（knowledge/tool/mcp）→ 等 RAG/工具域。
- 智能体编排（SAA Graph：编排引擎 + router/sequential/supervisor/group chat 多智能体模式）→ 已登记 specs/README"规划中的后续模块"，届时 ai_agent 表扩展编排配置列。
- 列表后端分页 → 量级大再说（现全量 + 前端筛选）。
