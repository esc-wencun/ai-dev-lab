// 字典类型/字典数据/参数管理数据访问（对位 SysDictTypeMapper/SysDictDataMapper/SysConfigMapper.xml）。
package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// DictTypeDAO 字典类型。
type DictTypeDAO struct{ db *gorm.DB }

func NewDictTypeDAO(db *gorm.DB) *DictTypeDAO { return &DictTypeDAO{db: db} }

const dictTypeColumns = `select dt.dict_id, dt.dict_name, dt.dict_type, dt.status, dt.create_by, dt.create_time, dt.update_by, dt.update_time, dt.remark from sys_dict_type dt`

// DictTypeDO 字典类型。
type DictTypeDO struct {
	DictID     int64  `gorm:"column:dict_id" json:"dictId"`
	DictName   string `gorm:"column:dict_name" json:"dictName"`
	DictType   string `gorm:"column:dict_type" json:"dictType"`
	Status     string `gorm:"column:status" json:"status"`
	CreateBy   string `gorm:"column:create_by" json:"createBy"`
	UpdateBy   string `gorm:"column:update_by" json:"updateBy"`
	UpdateTime string `gorm:"column:update_time" json:"updateTime"`
	Remark     string `gorm:"column:remark" json:"remark"`
	CreateTime string `gorm:"column:create_time" json:"createTime"`
}

// SelectDictTypeList 条件列表（dictName/dictType like、status、日期区间在 handler 拼 params——当前按常用三条件）。
func (d *DictTypeDAO) SelectDictTypeList(ctx context.Context, dictName, dictType, status string) ([]DictTypeDO, error) {
	where := "where 1=1"
	var args []any
	if dictName != "" {
		where += " AND dt.dict_name like concat('%', ?, '%')"
		args = append(args, dictName)
	}
	if dictType != "" {
		where += " AND dt.dict_type like concat('%', ?, '%')"
		args = append(args, dictType)
	}
	if status != "" {
		where += " AND dt.status = ?"
		args = append(args, status)
	}
	var list []DictTypeDO
	err := d.db.WithContext(ctx).Raw(dictTypeColumns+" "+where+" order by dt.dict_id", args...).Scan(&list).Error
	return list, err
}

// SelectDictTypeById 按 ID。
func (d *DictTypeDAO) SelectDictTypeById(ctx context.Context, dictID int64) (*DictTypeDO, error) {
	var m DictTypeDO
	err := d.db.WithContext(ctx).Raw(dictTypeColumns+" where dt.dict_id = ?", dictID).Scan(&m).Error
	if err != nil || m.DictID == 0 {
		return nil, err
	}
	return &m, nil
}

// CheckDictTypeUnique 类型唯一。
func (d *DictTypeDAO) CheckDictTypeUnique(ctx context.Context, dictType string) (*DictTypeDO, error) {
	var m DictTypeDO
	err := d.db.WithContext(ctx).Raw(dictTypeColumns+" where dt.dict_type = ? limit 1", dictType).Scan(&m).Error
	if err != nil || m.DictID == 0 {
		return nil, err
	}
	return &m, nil
}

// InsertDictType 新增。
func (d *DictTypeDAO) InsertDictType(ctx context.Context, m *DictTypeDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_dict_type (dict_name, dict_type, status, create_by, create_time, remark) values (?, ?, ?, ?, now(), ?)",
		m.DictName, m.DictType, m.Status, m.CreateBy, m.Remark).Error
}

// UpdateDictType 修改。
func (d *DictTypeDAO) UpdateDictType(ctx context.Context, m *DictTypeDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_dict_type set dict_name = ?, dict_type = ?, status = ?, update_by = ?, update_time = now(), remark = ? where dict_id = ?",
		m.DictName, m.DictType, m.Status, m.UpdateBy, m.Remark, m.DictID).Error
}

// UpdateDictDataType 联动更新字典数据的类型名（对位 updateDictDataType）。
func (d *DictTypeDAO) UpdateDictDataType(ctx context.Context, oldType, newType string) error {
	return d.db.WithContext(ctx).Exec("update sys_dict_data set dict_type = ? where dict_type = ?", newType, oldType).Error
}

// CountDictDataByType 类型下数据量（删除保护）。
func (d *DictTypeDAO) CountDictDataByType(ctx context.Context, dictType string) (int64, error) {
	var n int64
	err := d.db.WithContext(ctx).Raw("select count(1) from sys_dict_data where dict_type = ?", dictType).Scan(&n).Error
	return n, err
}

// DeleteDictTypeById 物理删。
func (d *DictTypeDAO) DeleteDictTypeById(ctx context.Context, dictID int64) error {
	return d.db.WithContext(ctx).Exec("delete from sys_dict_type where dict_id = ?", dictID).Error
}

// DictDataDAO 字典数据。
type DictDataDAO struct{ db *gorm.DB }

func NewDictDataDAO(db *gorm.DB) *DictDataDAO { return &DictDataDAO{db: db} }

const dictDataColumns = `select dd.dict_code, dd.dict_sort, dd.dict_label, dd.dict_value, dd.dict_type, dd.css_class, dd.list_class, dd.is_default, dd.status, dd.create_by, dd.create_time, dd.remark from sys_dict_data dd`

// DictDataDO 字典数据。
type DictDataDO struct {
	DictCode   int64  `gorm:"column:dict_code" json:"dictCode"`
	DictSort   int    `gorm:"column:dict_sort" json:"dictSort"`
	DictLabel  string `gorm:"column:dict_label" json:"dictLabel"`
	DictValue  string `gorm:"column:dict_value" json:"dictValue"`
	DictType   string `gorm:"column:dict_type" json:"dictType"`
	CssClass   string `gorm:"column:css_class" json:"cssClass"`
	ListClass  string `gorm:"column:list_class" json:"listClass"`
	IsDefault  string `gorm:"column:is_default" json:"isDefault"`
	Status     string `gorm:"column:status" json:"status"`
	CreateBy   string `gorm:"column:create_by" json:"createBy"`
	UpdateTime string `gorm:"column:update_time" json:"updateTime"`
	UpdateBy   string `gorm:"column:update_by" json:"updateBy"`
	CreateTime string `gorm:"column:create_time" json:"createTime"`
	Remark     string `gorm:"column:remark" json:"remark"`
}

// SelectDictDataList 条件列表（dictType 相等 + dictLabel like + status）。
func (d *DictDataDAO) SelectDictDataList(ctx context.Context, dictType, dictLabel, status string) ([]DictDataDO, error) {
	where := "where 1=1"
	var args []any
	if dictType != "" {
		where += " AND dd.dict_type = ?"
		args = append(args, dictType)
	}
	if dictLabel != "" {
		where += " AND dd.dict_label like concat('%', ?, '%')"
		args = append(args, dictLabel)
	}
	if status != "" {
		where += " AND dd.status = ?"
		args = append(args, status)
	}
	var list []DictDataDO
	err := d.db.WithContext(ctx).Raw(dictDataColumns+" "+where+" order by dd.dict_type, dd.dict_sort", args...).Scan(&list).Error
	return list, err
}

// SelectDictDataByType 前端取字典（对位 selectDictDataByType：status=0 按排序）。
func (d *DictDataDAO) SelectDictDataByType(ctx context.Context, dictType string) ([]DictDataDO, error) {
	var list []DictDataDO
	err := d.db.WithContext(ctx).Raw(dictDataColumns+" where dd.status = '0' and dd.dict_type = ? order by dd.dict_sort asc", dictType).Scan(&list).Error
	return list, err
}

// SelectDictDataById 按 ID。
func (d *DictDataDAO) SelectDictDataById(ctx context.Context, dictCode int64) (*DictDataDO, error) {
	var m DictDataDO
	err := d.db.WithContext(ctx).Raw(dictDataColumns+" where dd.dict_code = ?", dictCode).Scan(&m).Error
	if err != nil || m.DictCode == 0 {
		return nil, err
	}
	return &m, nil
}

// InsertDictData / UpdateDictData / DeleteDictDataByIds。
func (d *DictDataDAO) InsertDictData(ctx context.Context, m *DictDataDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_dict_data (dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, create_by, create_time, remark) values (?, ?, ?, ?, ?, ?, ?, ?, ?, now(), ?)",
		m.DictSort, m.DictLabel, m.DictValue, m.DictType, m.CssClass, m.ListClass, m.IsDefault, m.Status, m.CreateBy, m.Remark).Error
}

func (d *DictDataDAO) UpdateDictData(ctx context.Context, m *DictDataDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_dict_data set dict_sort = ?, dict_label = ?, dict_value = ?, dict_type = ?, css_class = ?, list_class = ?, is_default = ?, status = ?, update_by = ?, update_time = now(), remark = ? where dict_code = ?",
		m.DictSort, m.DictLabel, m.DictValue, m.DictType, m.CssClass, m.ListClass, m.IsDefault, m.Status, m.UpdateBy, m.Remark, m.DictCode).Error
}

func (d *DictDataDAO) DeleteDictDataByIds(ctx context.Context, codes []int64) error {
	if len(codes) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(codes)), ",")
	args := make([]any, len(codes))
	for i, c := range codes {
		args[i] = c
	}
	return d.db.WithContext(ctx).Exec("delete from sys_dict_data where dict_code in ("+ph+")", args...).Error
}

// ConfigDAO 参数管理。
type ConfigDAO struct{ db *gorm.DB }

func NewConfigDAO(db *gorm.DB) *ConfigDAO { return &ConfigDAO{db: db} }

const configColumns = `select c.config_id, c.config_name, c.config_key, c.config_value, c.config_type, c.create_by, c.create_time, c.update_by, c.update_time, c.remark from sys_config c`

// ConfigDO 参数。
type ConfigDO struct {
	ConfigID    int64  `gorm:"column:config_id" json:"configId"`
	ConfigName  string `gorm:"column:config_name" json:"configName"`
	ConfigKey   string `gorm:"column:config_key" json:"configKey"`
	ConfigValue string `gorm:"column:config_value" json:"configValue"`
	ConfigType  string `gorm:"column:config_type" json:"configType"`
	CreateBy    string `gorm:"column:create_by" json:"createBy"`
	CreateTime  string `gorm:"column:create_time" json:"createTime"`
	UpdateBy    string `gorm:"column:update_by" json:"updateBy"`
	UpdateTime  string `gorm:"column:update_time" json:"updateTime"`
	Remark      string `gorm:"column:remark" json:"remark"`
}

// SelectConfigList 条件列表（configName/configKey like、configType）。
func (d *ConfigDAO) SelectConfigList(ctx context.Context, configName, configKey, configType string) ([]ConfigDO, error) {
	where := "where 1=1"
	var args []any
	if configName != "" {
		where += " AND c.config_name like concat('%', ?, '%')"
		args = append(args, configName)
	}
	if configKey != "" {
		where += " AND c.config_key like concat('%', ?, '%')"
		args = append(args, configKey)
	}
	if configType != "" {
		where += " AND c.config_type = ?"
		args = append(args, configType)
	}
	var list []ConfigDO
	err := d.db.WithContext(ctx).Raw(configColumns+" "+where+" order by c.config_id", args...).Scan(&list).Error
	return list, err
}

// SelectConfigById 按 ID。
func (d *ConfigDAO) SelectConfigById(ctx context.Context, configID int64) (*ConfigDO, error) {
	var m ConfigDO
	err := d.db.WithContext(ctx).Raw(configColumns+" where c.config_id = ?", configID).Scan(&m).Error
	if err != nil || m.ConfigID == 0 {
		return nil, err
	}
	return &m, nil
}

// CheckConfigKeyUnique 键唯一。
func (d *ConfigDAO) CheckConfigKeyUnique(ctx context.Context, key string) (*ConfigDO, error) {
	var m ConfigDO
	err := d.db.WithContext(ctx).Raw(configColumns+" where c.config_key = ? limit 1", key).Scan(&m).Error
	if err != nil || m.ConfigID == 0 {
		return nil, err
	}
	return &m, nil
}

// InsertConfig / UpdateConfig / DeleteConfigByIds。
func (d *ConfigDAO) InsertConfig(ctx context.Context, m *ConfigDO) error {
	return d.db.WithContext(ctx).Exec(
		"insert into sys_config (config_name, config_key, config_value, config_type, create_by, create_time, remark) values (?, ?, ?, ?, ?, now(), ?)",
		m.ConfigName, m.ConfigKey, m.ConfigValue, m.ConfigType, m.CreateBy, m.Remark).Error
}

func (d *ConfigDAO) UpdateConfig(ctx context.Context, m *ConfigDO) error {
	return d.db.WithContext(ctx).Exec(
		"update sys_config set config_name = ?, config_key = ?, config_value = ?, config_type = ?, update_by = ?, update_time = now(), remark = ? where config_id = ?",
		m.ConfigName, m.ConfigKey, m.ConfigValue, m.ConfigType, m.UpdateBy, m.Remark, m.ConfigID).Error
}

func (d *ConfigDAO) DeleteConfigByIds(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return d.db.WithContext(ctx).Exec("delete from sys_config where config_id in ("+ph+")", args...).Error
}

// DB 暴露底层句柄（configSelectValue 轻量查询）。
func (d *ConfigDAO) DB() *gorm.DB { return d.db }
