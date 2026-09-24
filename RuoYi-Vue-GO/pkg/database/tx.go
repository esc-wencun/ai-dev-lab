// Package database 数据库事务封装与 context 传播约定（对位 Java @Transactional / Python get_db）。
//
// context 传播约定（全局纪律，评审项）：
//  1. dao 层所有方法第一个参数收 ctx context.Context，来自 handler 的 c.Request.Context()；
//  2. GORM 一律 db.WithContext(ctx)、go-redis 一律 WithContext(ctx)——超时与取消全链路生效；
//  3. 禁止把带请求 ctx 的 *gorm.DB 实例存入全局变量或跨请求复用（对位 get_db 的会话泄漏检查）；
//     每次使用 db.WithContext(ctx) 得到的都是独立实例，天然无泄漏。
//
// 事务约定：
//   - 多表写入（用户+关联表、角色+菜单等）必须包 InTx 显式事务；
//   - 单表写由 service 自行决定是否需要事务；
//   - fn 内所有读写必须用 tx（已 WithContext），不得逃逸回外层 db。
package database

import (
	"context"

	"gorm.io/gorm"
)

// InTx 显式事务唯一入口：fn 返回 error（含 *errors.BusinessError）整体回滚并原样透传错误，
// 返回 nil 提交。错误类型不丢失——上层 middleware.Wrap 依赖它映射响应信封。
func InTx(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}
