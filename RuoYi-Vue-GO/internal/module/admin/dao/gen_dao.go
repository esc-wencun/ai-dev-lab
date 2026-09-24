// 代码生成器数据访问（对位 GenTableMapper/GenTableColumnMapper；降级范围见 spec）。
package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// GenTableDAO 代码生成业务表。
type GenTableDAO struct{ db *gorm.DB }

func NewGenTableDAO(db *gorm.DB) *GenTableDAO { return &GenTableDAO{db: db} }

// GenTableDO 生成表信息（对位 GenTable 的核心字段）。
type GenTableDO struct {
	TableID        int64         `gorm:"column:table_id" json:"tableId"`
	TableName      string        `gorm:"column:table_name" json:"tableName"`
	TableComment   string        `gorm:"column:table_comment" json:"tableComment"`
	ClassName      string        `gorm:"column:class_name" json:"className"`
	PackageName    string        `gorm:"column:package_name" json:"packageName"`
	ModuleName     string        `gorm:"column:module_name" json:"moduleName"`
	BusinessName   string        `gorm:"column:business_name" json:"businessName"`
	FunctionName   string        `gorm:"column:function_name" json:"functionName"`
	FunctionAuthor string        `gorm:"column:function_author" json:"functionAuthor"`
	CreateTime     string        `gorm:"column:create_time" json:"createTime"`
	UpdateTime     string        `gorm:"column:update_time" json:"updateTime"`
	Remark         string        `gorm:"column:remark" json:"remark"`
	Columns        []GenColumnDO `gorm:"-" json:"columns,omitempty"`
}

// GenColumnDO 生成列信息。
type GenColumnDO struct {
	ColumnID      int64  `gorm:"column:column_id" json:"columnId"`
	TableID       int64  `gorm:"column:table_id" json:"tableId"`
	ColumnName    string `gorm:"column:column_name" json:"columnName"`
	ColumnComment string `gorm:"column:column_comment" json:"columnComment"`
	ColumnType    string `gorm:"column:column_type" json:"columnType"`
	Sort          int    `gorm:"column:sort" json:"sort"`
	IsPK          string `gorm:"column:is_pk" json:"isPk"`
	IsIncrement   string `gorm:"column:is_increment" json:"isIncrement"`
	IsRequired    string `gorm:"column:is_required" json:"isRequired"`
	IsInsert      string `gorm:"column:is_insert" json:"isInsert"`
	IsEdit        string `gorm:"column:is_edit" json:"isEdit"`
	IsList        string `gorm:"column:is_list" json:"isList"`
	IsQuery       string `gorm:"column:is_query" json:"isQuery"`
	QueryType     string `gorm:"column:query_type" json:"queryType"`
	HtmlType      string `gorm:"column:html_type" json:"htmlType"`
	DictType      string `gorm:"column:dict_type" json:"dictType"`
	Sort_         int    `gorm:"column:sort" json:"-"`
}

// SelectGenList 条件列表（tableName/tableComment like）。
func (d *GenTableDAO) SelectGenList(ctx context.Context, tableName, tableComment string) ([]GenTableDO, error) {
	where := "where 1=1"
	var args []any
	if tableName != "" {
		where += " AND table_name like concat('%', ?, '%')"
		args = append(args, tableName)
	}
	if tableComment != "" {
		where += " AND table_comment like concat('%', ?, '%')"
		args = append(args, tableComment)
	}
	var list []GenTableDO
	err := d.db.WithContext(ctx).Raw("select table_id, table_name, table_comment, class_name, package_name, module_name, business_name, function_name, function_author, create_time, update_time, remark from gen_table "+where+" order by table_id", args...).Scan(&list).Error
	return list, err
}

// SelectDbTables 库中未导入的表（对位 selectDbTableList：读 information_schema）。
func (d *GenTableDAO) SelectDbTables(ctx context.Context, tableName, tableComment string) ([]GenTableDO, error) {
	where := `where table_schema = database() and table_name not in (select table_name from gen_table)
		and table_name not like 'gen_%' and table_name not in ('sys_dict_type','sys_dict_data','sys_config','sys_job','sys_job_log','sys_logininfor','sys_oper_log','sys_notice','sys_notice_read')`
	var args []any
	if tableName != "" {
		where += " and table_name like concat('%', ?, '%')"
		args = append(args, tableName)
	}
	if tableComment != "" {
		where += " and table_comment like concat('%', ?, '%')"
		args = append(args, tableComment)
	}
	var list []GenTableDO
	err := d.db.WithContext(ctx).Raw(`select table_name, table_comment, create_time from information_schema.tables
		`+where, args...).Scan(&list).Error
	return list, err
}

// SelectGenTableById 详情（含列）。
func (d *GenTableDAO) SelectGenTableById(ctx context.Context, tableID int64) (*GenTableDO, error) {
	var m GenTableDO
	err := d.db.WithContext(ctx).Raw("select table_id, table_name, table_comment, class_name, package_name, module_name, business_name, function_name, function_author, create_time, update_time, remark from gen_table where table_id = ?", tableID).Scan(&m).Error
	if err != nil || m.TableID == 0 {
		return nil, err
	}
	m.Columns, _ = d.SelectColumnsByTable(ctx, tableID)
	return &m, nil
}

// SelectColumnsByTable 列信息。
func (d *GenTableDAO) SelectColumnsByTable(ctx context.Context, tableID int64) ([]GenColumnDO, error) {
	var list []GenColumnDO
	err := d.db.WithContext(ctx).Raw("select column_id, table_id, column_name, column_comment, column_type, sort, is_pk, is_increment, is_required, is_insert, is_edit, is_list, is_query, query_type, html_type, dict_type from gen_table_column where table_id = ? order by sort", tableID).Scan(&list).Error
	return list, err
}

// InsertGenTable 导入表（主表）。
func (d *GenTableDAO) InsertGenTable(ctx context.Context, m *GenTableDO) (int64, error) {
	res := d.db.WithContext(ctx).Exec(
		"insert into gen_table (table_name, table_comment, class_name, package_name, module_name, business_name, function_name, function_author, create_time, remark) values (?, ?, ?, ?, ?, ?, ?, ?, now(), ?)",
		m.TableName, m.TableComment, m.ClassName, m.PackageName, m.ModuleName, m.BusinessName, m.FunctionName, m.FunctionAuthor, m.Remark)
	if res.Error != nil {
		return 0, res.Error
	}
	var id int64
	d.db.WithContext(ctx).Raw("select table_id from gen_table where table_name = ? order by table_id desc limit 1", m.TableName).Scan(&id)
	return id, nil
}

// InsertGenColumns 从 information_schema 导入列（对位 GenTableServiceImpl 的列装配）。
func (d *GenTableDAO) InsertGenColumns(ctx context.Context, tableID int64, tableName string) error {
	var cols []struct {
		ColumnName    string `gorm:"column:column_name"`
		ColumnComment string `gorm:"column:column_comment"`
		ColumnType    string `gorm:"column:column_type"`
		ColumnKey     string `gorm:"column:column_key"`
		Extra         string `gorm:"column:extra"`
		IsNullable    string `gorm:"column:is_nullable"`
		Ordinal       int    `gorm:"column:ordinal_position"`
	}
	if err := d.db.WithContext(ctx).Raw(`select column_name, column_comment, column_type, column_key, extra, is_nullable, ordinal_position
		from information_schema.columns where table_schema = database() and table_name = ? order by ordinal_position`, tableName).Scan(&cols).Error; err != nil {
		return err
	}
	for i, c := range cols {
		isPK := "0"
		if c.ColumnKey == "PRI" {
			isPK = "1"
		}
		isInc := "0"
		if strings.Contains(c.Extra, "auto_increment") {
			isInc = "1"
		}
		isReq := "0"
		if c.IsNullable == "NO" && isPK == "0" {
			isReq = "1"
		}
		if err := d.db.WithContext(ctx).Exec(
			"insert into gen_table_column (table_id, column_name, column_comment, column_type, sort, is_pk, is_increment, is_required, is_insert, is_edit, is_list, is_query, query_type, html_type, dict_type) values (?, ?, ?, ?, ?, ?, ?, ?, '1', '1', '1', '0', 'EQ', 'input', '')",
			tableID, c.ColumnName, c.ColumnComment, strings.ToUpper(c.ColumnType), i+1, isPK, isInc, isReq).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteGenTableByIds 批量删（主表+列）。
func (d *GenTableDAO) DeleteGenTableByIds(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("delete from gen_table_column where table_id in ("+ph+")", args...).Error; err != nil {
			return err
		}
		return tx.Exec("delete from gen_table where table_id in ("+ph+")", args...).Error
	})
}

var _ = gorm.ErrRecordNotFound
