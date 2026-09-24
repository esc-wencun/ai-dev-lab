package database

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	bizerrors "ruoyi-vue-go/internal/common/errors"
)

// 模拟两张关联表（对位"多表写入必须显式事务"的场景：主表+子表）
type orderDO struct {
	ID     int64 `gorm:"primaryKey"`
	Amount int   `json:"amount"`
}

type itemDO struct {
	ID      int64  `gorm:"primaryKey"`
	OrderID int64  `json:"orderId"`
	Sku     string `json:"sku"`
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 sqlite 内存库失败: %v", err)
	}
	if err := db.AutoMigrate(&orderDO{}, &itemDO{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	return db
}

func count(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()
	var n int64
	if err := db.Model(model).Count(&n).Error; err != nil {
		t.Fatalf("count 失败: %v", err)
	}
	return n
}

// TestInTxCommit fn 返回 nil → 两表写入全部提交
func TestInTxCommit(t *testing.T) {
	db := newTestDB(t)
	err := InTx(context.Background(), db, func(tx *gorm.DB) error {
		if err := tx.Create(&orderDO{ID: 1, Amount: 100}).Error; err != nil {
			return err
		}
		return tx.Create(&itemDO{ID: 1, OrderID: 1, Sku: "sku-a"}).Error
	})
	if err != nil {
		t.Fatalf("事务应提交成功: %v", err)
	}
	if got := count(t, db, &orderDO{}); got != 1 {
		t.Errorf("主表应有 1 条，实际 %d", got)
	}
	if got := count(t, db, &itemDO{}); got != 1 {
		t.Errorf("子表应有 1 条，实际 %d", got)
	}
}

// TestInTxRollbackOnPlainError service 返回普通 error → 多表写入全部回滚
func TestInTxRollbackOnPlainError(t *testing.T) {
	db := newTestDB(t)
	boom := errors.New("下游服务失败")
	err := InTx(context.Background(), db, func(tx *gorm.DB) error {
		if err := tx.Create(&orderDO{ID: 1, Amount: 100}).Error; err != nil {
			return err
		}
		if err := tx.Create(&itemDO{ID: 1, OrderID: 1, Sku: "sku-a"}).Error; err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("错误应原样透传: %v", err)
	}
	if got := count(t, db, &orderDO{}); got != 0 {
		t.Errorf("回滚后主表应为空，实际 %d 条", got)
	}
	if got := count(t, db, &itemDO{}); got != 0 {
		t.Errorf("回滚后子表应为空，实际 %d 条", got)
	}
}

// TestInTxRollbackKeepsBusinessError fn 返回 *BusinessError → 回滚且错误类型不丢失
// （middleware.Wrap 依赖类型映射响应信封，事务层不得吞掉类型）
func TestInTxRollbackKeepsBusinessError(t *testing.T) {
	db := newTestDB(t)
	be := bizerrors.NewWithCode(601, "库存不足")
	err := InTx(context.Background(), db, func(tx *gorm.DB) error {
		if err := tx.Create(&orderDO{ID: 2, Amount: 1}).Error; err != nil {
			return err
		}
		return be
	})

	var got *bizerrors.BusinessError
	if !errors.As(err, &got) {
		t.Fatalf("BusinessError 类型应保留，实际 %T: %v", err, err)
	}
	if got.Code != 601 || got.Msg != "库存不足" {
		t.Errorf("Code/Msg 不符: %+v", got)
	}
	if count(t, db, &orderDO{}) != 0 {
		t.Error("BusinessError 触发的回滚后主表应为空")
	}
}

// TestInTxAutoRollbackOnPanic fn panic → GORM 自动回滚并向上抛出
func TestInTxAutoRollbackOnPanic(t *testing.T) {
	db := newTestDB(t)
	func() {
		defer func() {
			if rec := recover(); rec == nil {
				t.Error("panic 应向上传播")
			}
		}()
		_ = InTx(context.Background(), db, func(tx *gorm.DB) error {
			if err := tx.Create(&orderDO{ID: 3, Amount: 1}).Error; err != nil {
				return err
			}
			panic("意外崩溃")
		})
	}()
	if got := count(t, db, &orderDO{}); got != 0 {
		t.Errorf("panic 后主表应回滚为空，实际 %d 条", got)
	}
}
