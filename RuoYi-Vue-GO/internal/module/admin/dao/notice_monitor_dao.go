// 通知公告 + 监控日志数据访问（对位 SysNoticeMapper/SysNoticeReadMapper/SysOperLogMapper/SysLogininforMapper.xml）。
package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// NoticeDAO 通知公告。
type NoticeDAO struct{ db *gorm.DB }

func NewNoticeDAO(db *gorm.DB) *NoticeDAO { return &NoticeDAO{db: db} }

const noticeColumns = `select n.notice_id, n.notice_title, n.notice_type, n.notice_content, n.status, n.create_by, n.create_time, n.update_by, n.update_time, n.remark from sys_notice n`

// NoticeDO 公告。
type NoticeDO struct {
	NoticeID      int64  `gorm:"column:notice_id" json:"noticeId"`
	NoticeTitle   string `gorm:"column:notice_title" json:"noticeTitle"`
	NoticeType    string `gorm:"column:notice_type" json:"noticeType"`
	NoticeContent string `gorm:"column:notice_content" json:"noticeContent"`
	Status        string `gorm:"column:status" json:"status"`
	CreateBy      string `gorm:"column:create_by" json:"createBy"`
	CreateTime    string `gorm:"column:create_time" json:"createTime"`
	UpdateBy      string `gorm:"column:update_by" json:"updateBy"`
	UpdateTime    string `gorm:"column:update_time" json:"updateTime"`
	Remark        string `gorm:"column:remark" json:"remark"`
	IsRead        bool   `gorm:"-" json:"isRead"`
}

// SelectNoticeList 条件列表（title like/type/createBy like）。
func (d *NoticeDAO) SelectNoticeList(ctx context.Context, title, noticeType, createBy string) ([]NoticeDO, error) {
	where := "where 1=1"
	var args []any
	if title != "" {
		where += " AND n.notice_title like concat('%', ?, '%')"
		args = append(args, title)
	}
	if noticeType != "" {
		where += " AND n.notice_type = ?"
		args = append(args, noticeType)
	}
	if createBy != "" {
		where += " AND n.create_by like concat('%', ?, '%')"
		args = append(args, createBy)
	}
	var list []NoticeDO
	err := d.db.WithContext(ctx).Raw(noticeColumns+" "+where+" order by n.notice_id desc", args...).Scan(&list).Error
	return list, err
}

// SelectNoticeById 按 ID。
func (d *NoticeDAO) SelectNoticeById(ctx context.Context, id int64) (*NoticeDO, error) {
	var m NoticeDO
	err := d.db.WithContext(ctx).Raw(noticeColumns+" where n.notice_id = ?", id).Scan(&m).Error
	if err != nil || m.NoticeID == 0 {
		return nil, err
	}
	return &m, nil
}

// InsertNotice / UpdateNotice / DeleteNoticeByIds。
func (d *NoticeDAO) InsertNotice(ctx context.Context, m *NoticeDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_notice (notice_title, notice_type, notice_content, status, create_by, create_time, remark) values (?, ?, ?, ?, ?, now(), ?)",
		m.NoticeTitle, m.NoticeType, m.NoticeContent, m.Status, m.CreateBy, m.Remark).Error
}

func (d *NoticeDAO) UpdateNotice(ctx context.Context, m *NoticeDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_notice set notice_title = ?, notice_type = ?, notice_content = ?, status = ?, update_by = ?, update_time = now(), remark = ? where notice_id = ?",
		m.NoticeTitle, m.NoticeType, m.NoticeContent, m.Status, m.UpdateBy, m.Remark, m.NoticeID).Error
}

func (d *NoticeDAO) DeleteNoticeByIds(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return d.db.WithContext(ctx).Exec("delete from sys_notice where notice_id in ("+ph+")", args...).Error
}

// ListTopWithReadStatus 首页公告（对位 selectNoticeListWithReadStatus）。
func (d *NoticeDAO) ListTopWithReadStatus(ctx context.Context, userID int64, limit int) ([]NoticeDO, error) {
	var list []NoticeDO
	err := d.db.WithContext(ctx).Raw(`select
			n.notice_id    as notice_id,
			n.notice_title as notice_title,
			n.notice_type  as notice_type,
			n.status,
			n.create_by    as create_by,
			n.create_time  as create_time,
			case when r.notice_id is not null then true else false end as is_read
		from sys_notice n
		left join sys_notice_read r
			on r.notice_id = n.notice_id and r.user_id = ?
		where n.status = '0'
		order by n.notice_id desc
		limit ?`, userID, limit).Scan(&list).Error
	return list, err
}

// MarkRead 标记已读（唯一键冲突忽略）。
func (d *NoticeDAO) MarkRead(ctx context.Context, userID, noticeID int64) error {
	var n int64
	d.db.WithContext(ctx).Raw("select count(1) from sys_notice_read where user_id = ? and notice_id = ?", userID, noticeID).Scan(&n)
	if n == 0 {
		return d.db.WithContext(ctx).Exec("insert into sys_notice_read (notice_id, user_id, read_time) values (?, ?, now())", noticeID, userID).Error
	}
	return nil
}

// ReadUsersByNotice 已读用户列表。
func (d *NoticeDAO) ReadUsersByNotice(ctx context.Context, noticeID int64, search string) ([]map[string]any, error) {
	where := `select u.user_id as userId, u.nick_name as nickName, u.user_name as userName, r.read_time as readTime
		from sys_notice_read r join sys_user u on u.user_id = r.user_id where r.notice_id = ?`
	var args []any
	args = append(args, noticeID)
	if search != "" {
		where += " AND (u.nick_name like concat('%', ?, '%') OR u.user_name like concat('%', ?, '%'))"
		args = append(args, search, search)
	}
	var list []map[string]any
	err := d.db.WithContext(ctx).Raw(where, args...).Scan(&list).Error
	return list, err
}

// OperLogDAO 操作日志查询（写入在 aspect 包）。
type OperLogDAO struct{ db *gorm.DB }

func NewOperLogDAO(db *gorm.DB) *OperLogDAO { return &OperLogDAO{db: db} }

// SelectOperLogList 条件列表（title/operName/businessType/status/timeRange like-eq）。
func (d *OperLogDAO) SelectOperLogList(ctx context.Context, title, operName, businessType, status string) ([]map[string]any, error) {
	where := "where 1=1"
	var args []any
	if title != "" {
		where += " AND title like concat('%', ?, '%')"
		args = append(args, title)
	}
	if operName != "" {
		where += " AND oper_name like concat('%', ?, '%')"
		args = append(args, operName)
	}
	if businessType != "" && businessType != "-1" {
		where += " AND business_type = ?"
		args = append(args, businessType)
	}
	if status != "" && status != "-1" {
		where += " AND status = ?"
		args = append(args, status)
	}
	var list []map[string]any
	err := d.db.WithContext(ctx).Raw("select * from sys_oper_log "+where+" order by oper_id desc", args...).Scan(&list).Error
	return list, err
}

// DeleteOperLogByIds / CleanOperLog。
func (d *OperLogDAO) DeleteOperLogByIds(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return d.db.WithContext(ctx).Exec("delete from sys_oper_log where oper_id in ("+ph+")", args...).Error
}

func (d *OperLogDAO) CleanOperLog(ctx context.Context) error {
	return d.db.WithContext(ctx).Exec("truncate table sys_oper_log").Error
}

// LogininforDAO 登录日志。
type LogininforDAO struct{ db *gorm.DB }

func NewLogininforDAO(db *gorm.DB) *LogininforDAO { return &LogininforDAO{db: db} }

// SelectLogininforList 条件列表（userName/ipaddr like、status、timeRange）。
func (d *LogininforDAO) SelectLogininforList(ctx context.Context, userName, ipaddr, status string) ([]map[string]any, error) {
	where := "where 1=1"
	var args []any
	if userName != "" {
		where += " AND user_name like concat('%', ?, '%')"
		args = append(args, userName)
	}
	if ipaddr != "" {
		where += " AND ipaddr like concat('%', ?, '%')"
		args = append(args, ipaddr)
	}
	if status != "" && status != "-1" {
		where += " AND status = ?"
		args = append(args, status)
	}
	var list []map[string]any
	err := d.db.WithContext(ctx).Raw("select * from sys_logininfor "+where+" order by info_id desc", args...).Scan(&list).Error
	return list, err
}

// DeleteLogininforByIds / CleanLogininfor。
func (d *LogininforDAO) DeleteLogininforByIds(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return d.db.WithContext(ctx).Exec("delete from sys_logininfor where info_id in ("+ph+")", args...).Error
}

func (d *LogininforDAO) CleanLogininfor(ctx context.Context) error {
	return d.db.WithContext(ctx).Exec("truncate table sys_logininfor").Error
}

// UnlockUser 解锁：清 pwd_err_cnt（对位 unlock；Redis 操作在 service 层做，这里仅为接口聚合注释）。

// OnlineUser 在线用户（对位 SysUserOnline：从 Redis 会话聚合；无 DB 查询）。
type OnlineUserDAO struct{ db *gorm.DB }

func NewOnlineUserDAO(db *gorm.DB) *OnlineUserDAO { return &OnlineUserDAO{db: db} }

var _ = gorm.ErrRecordNotFound
