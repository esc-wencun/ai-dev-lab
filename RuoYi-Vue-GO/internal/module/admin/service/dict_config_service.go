// 字典 + 参数管理业务（对位 SysDictTypeServiceImpl/SysDictDataServiceImpl/SysConfigServiceImpl 的缓存联动）。
package service

import (
	"context"
	"strings"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/common/constant"
	"ruoyi-vue-go/internal/module/admin/dao"
)

// DictTypeService 字典类型。
type DictTypeService struct {
	dictType *dao.DictTypeDAO
	dictData *dao.DictDataDAO
	cache    *cache.RedisCache
}

func NewDictTypeService(dt *dao.DictTypeDAO, dd *dao.DictDataDAO, rc *cache.RedisCache) *DictTypeService {
	return &DictTypeService{dictType: dt, dictData: dd, cache: rc}
}

// List 列表。
func (s *DictTypeService) List(ctx context.Context, dictName, dictType, status string) ([]dao.DictTypeDO, error) {
	return s.dictType.SelectDictTypeList(ctx, dictName, dictType, status)
}

// GetByID 详情。
func (s *DictTypeService) GetByID(ctx context.Context, id int64) (*dao.DictTypeDO, error) {
	return s.dictType.SelectDictTypeById(ctx, id)
}

// Optionselect 全部类型（数据回显）。
func (s *DictTypeService) Optionselect(ctx context.Context) ([]dao.DictTypeDO, error) {
	return s.dictType.SelectDictTypeList(ctx, "", "", "")
}

// Input 字典类型参数。
type DictTypeInput struct {
	DictID   int64  `json:"dictId"`
	DictName string `json:"dictName"`
	DictType string `json:"dictType"`
	Status   string `json:"status"`
	Remark   string `json:"remark"`
}

// Add 新增（唯一校验 + 缓存失效）。
func (s *DictTypeService) Add(ctx context.Context, in *DictTypeInput, operator string) error {
	if info, _ := s.dictType.CheckDictTypeUnique(ctx, in.DictType); info != nil {
		return errf("新增字典'%s'失败，字典类型已存在", in.DictName)
	}
	m := &dao.DictTypeDO{DictName: in.DictName, DictType: in.DictType, Status: in.Status, Remark: in.Remark, CreateBy: operator}
	if err := s.dictType.InsertDictType(ctx, m); err != nil {
		return err
	}
	return s.refreshDictCache(ctx, in.DictType)
}

// Edit 修改（类型名变更联动字典数据 + 缓存刷新；对位 updateDictType）。
func (s *DictTypeService) Edit(ctx context.Context, in *DictTypeInput, operator string) error {
	old, err := s.dictType.SelectDictTypeById(ctx, in.DictID)
	if err != nil {
		return err
	}
	if info, _ := s.dictType.CheckDictTypeUnique(ctx, in.DictType); info != nil && info.DictID != in.DictID {
		return errf("修改字典'%s'失败，字典类型已存在", in.DictName)
	}
	if old.DictType != in.DictType {
		if err := s.dictType.UpdateDictDataType(ctx, old.DictType, in.DictType); err != nil {
			return err
		}
		_, _ = s.cache.Delete(ctx, constant.SysDictKey+old.DictType)
	}
	m := &dao.DictTypeDO{DictID: in.DictID, DictName: in.DictName, DictType: in.DictType, Status: in.Status, Remark: in.Remark, UpdateBy: operator}
	if err := s.dictType.UpdateDictType(ctx, m); err != nil {
		return err
	}
	return s.refreshDictCache(ctx, in.DictType)
}

// Remove 删除（数据保护 + 缓存清除；文案对位 %s已分配,不能删除）。
func (s *DictTypeService) Remove(ctx context.Context, ids []int64) error {
	for _, id := range ids {
		dt, err := s.dictType.SelectDictTypeById(ctx, id)
		if err != nil {
			continue
		}
		if n, _ := s.dictType.CountDictDataByType(ctx, dt.DictType); n > 0 {
			return errf("%s已分配,不能删除", dt.DictName)
		}
		_ = s.dictType.DeleteDictTypeById(ctx, id)
		_, _ = s.cache.Delete(ctx, constant.SysDictKey+dt.DictType)
	}
	return nil
}

// RefreshCache 刷新全部字典缓存（对位 refreshCache：删除全部 sys_dict: 键，首次读取回源）。
func (s *DictTypeService) RefreshCache(ctx context.Context) error {
	keys, err := s.cache.KeysByPrefix(ctx, constant.SysDictKey)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		_, err = s.cache.Delete(ctx, keys...)
	}
	return err
}

// refreshDictCache 预热某类型缓存（对位 DictUtils.setDictCache：存 sys_dict:{type}）。
func (s *DictTypeService) refreshDictCache(ctx context.Context, dictType string) error {
	datas, err := s.dictData.SelectDictDataByType(ctx, dictType)
	if err != nil {
		return err
	}
	return s.cache.SetObject(ctx, constant.SysDictKey+dictType, datas, 0)
}

// DictDataService 字典数据。
type DictDataService struct {
	dictData *dao.DictDataDAO
	cache    *cache.RedisCache
}

func NewDictDataService(dd *dao.DictDataDAO, rc *cache.RedisCache) *DictDataService {
	return &DictDataService{dictData: dd, cache: rc}
}

// List 列表（按类型/标签/状态）。
func (s *DictDataService) List(ctx context.Context, dictType, dictLabel, status string) ([]dao.DictDataDO, error) {
	return s.dictData.SelectDictDataList(ctx, dictType, dictLabel, status)
}

// GetByID 详情。
func (s *DictDataService) GetByID(ctx context.Context, dictCode int64) (*dao.DictDataDO, error) {
	return s.dictData.SelectDictDataById(ctx, dictCode)
}

// OptionType 按类型取字典（对位 /type/{dictType}：优先 Redis 缓存）。
func (s *DictDataService) OptionType(ctx context.Context, dictType string) ([]dao.DictDataDO, error) {
	var cached []dao.DictDataDO
	if err := s.cache.GetObject(ctx, constant.SysDictKey+dictType, &cached); err == nil && cached != nil {
		return cached, nil
	}
	list, err := s.dictData.SelectDictDataByType(ctx, dictType)
	if err != nil {
		return nil, err
	}
	_ = s.cache.SetObject(ctx, constant.SysDictKey+dictType, list, 0)
	return list, nil
}

// DictDataInput 字典数据参数。
type DictDataInput struct {
	DictCode  int64  `json:"dictCode"`
	DictSort  int    `json:"dictSort"`
	DictLabel string `json:"dictLabel"`
	DictValue string `json:"dictValue"`
	DictType  string `json:"dictType"`
	CssClass  string `json:"cssClass"`
	ListClass string `json:"listClass"`
	IsDefault string `json:"isDefault"`
	Status    string `json:"status"`
	Remark    string `json:"remark"`
}

// Add 新增 + 缓存刷新。
func (s *DictDataService) Add(ctx context.Context, in *DictDataInput, operator string) error {
	m := &dao.DictDataDO{DictSort: in.DictSort, DictLabel: in.DictLabel, DictValue: in.DictValue, DictType: in.DictType,
		CssClass: in.CssClass, ListClass: in.ListClass, IsDefault: in.IsDefault, Status: in.Status, Remark: in.Remark, CreateBy: operator}
	if err := s.dictData.InsertDictData(ctx, m); err != nil {
		return err
	}
	return s.evict(ctx, in.DictType)
}

// Edit 修改 + 缓存刷新。
func (s *DictDataService) Edit(ctx context.Context, in *DictDataInput, operator string) error {
	m := &dao.DictDataDO{DictCode: in.DictCode, DictSort: in.DictSort, DictLabel: in.DictLabel, DictValue: in.DictValue,
		DictType: in.DictType, CssClass: in.CssClass, ListClass: in.ListClass, IsDefault: in.IsDefault,
		Status: in.Status, Remark: in.Remark, UpdateBy: operator}
	if err := s.dictData.UpdateDictData(ctx, m); err != nil {
		return err
	}
	return s.evict(ctx, in.DictType)
}

// Remove 批量删 + 缓存清除。
func (s *DictDataService) Remove(ctx context.Context, codes []int64) error {
	// 先取涉及类型用于清缓存
	types := map[string]bool{}
	for _, code := range codes {
		if m, err := s.dictData.SelectDictDataById(ctx, code); err == nil {
			types[m.DictType] = true
		}
	}
	if err := s.dictData.DeleteDictDataByIds(ctx, codes); err != nil {
		return err
	}
	for t := range types {
		_ = s.evict(ctx, t)
	}
	return nil
}

func (s *DictDataService) evict(ctx context.Context, dictType string) error {
	_, err := s.cache.Delete(ctx, constant.SysDictKey+dictType)
	return err
}

// ConfigService 参数管理。
type ConfigService struct {
	config *dao.ConfigDAO
	cache  *cache.RedisCache
}

func NewConfigService(c *dao.ConfigDAO, rc *cache.RedisCache) *ConfigService {
	return &ConfigService{config: c, cache: rc}
}

// List 列表。
func (s *ConfigService) List(ctx context.Context, configName, configKey, configType string) ([]dao.ConfigDO, error) {
	return s.config.SelectConfigList(ctx, configName, configKey, configType)
}

// GetByID 详情。
func (s *ConfigService) GetByID(ctx context.Context, id int64) (*dao.ConfigDO, error) {
	return s.config.SelectConfigById(ctx, id)
}

// ConfigKeyByKey 按键取值（对位 /configKey/{configKey}：缓存优先）。
func (s *ConfigService) ConfigKeyByKey(ctx context.Context, key string) (string, error) {
	var cached string
	if err := s.cache.GetObject(ctx, constant.SysConfigKey+key, &cached); err == nil && cached != "" {
		return cached, nil
	}
	var m struct {
		ConfigValue string `gorm:"column:config_value"`
	}
	if err := s.configSelectValue(ctx, key, &m); err != nil {
		return "", err
	}
	_ = s.cache.SetObject(ctx, constant.SysConfigKey+key, m.ConfigValue, 0)
	return m.ConfigValue, nil
}

func (s *ConfigService) configSelectValue(ctx context.Context, key string, dest any) error {
	return s.config.DB().Raw("select config_value from sys_config where config_key = ?", key).Scan(dest).Error
}

// ConfigInput 参数参数。
type ConfigInput struct {
	ConfigID    int64  `json:"configId"`
	ConfigName  string `json:"configName"`
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ConfigType  string `json:"configType"`
	Remark      string `json:"remark"`
}

// Add 新增（键唯一 `新增参数'x'失败，参数键名已存在`）。
func (s *ConfigService) Add(ctx context.Context, in *ConfigInput, operator string) error {
	if info, _ := s.config.CheckConfigKeyUnique(ctx, in.ConfigKey); info != nil {
		return errf("新增参数'%s'失败，参数键名已存在", in.ConfigName)
	}
	m := &dao.ConfigDO{ConfigName: in.ConfigName, ConfigKey: in.ConfigKey, ConfigValue: in.ConfigValue, ConfigType: in.ConfigType, Remark: in.Remark, CreateBy: operator}
	return s.config.InsertConfig(ctx, m)
}

// Edit 修改 + 缓存更新。
func (s *ConfigService) Edit(ctx context.Context, in *ConfigInput, operator string) error {
	if info, _ := s.config.CheckConfigKeyUnique(ctx, in.ConfigKey); info != nil && info.ConfigID != in.ConfigID {
		return errf("修改参数'%s'失败，参数键名已存在", in.ConfigName)
	}
	m := &dao.ConfigDO{ConfigID: in.ConfigID, ConfigName: in.ConfigName, ConfigKey: in.ConfigKey, ConfigValue: in.ConfigValue, ConfigType: in.ConfigType, Remark: in.Remark, UpdateBy: operator}
	if err := s.config.UpdateConfig(ctx, m); err != nil {
		return err
	}
	return s.cache.SetObject(ctx, constant.SysConfigKey+in.ConfigKey, in.ConfigValue, 0)
}

// Remove 批量删（内置保护）+ 缓存清除。
func (s *ConfigService) Remove(ctx context.Context, ids []int64) error {
	keys := map[string]bool{}
	for _, id := range ids {
		m, err := s.config.SelectConfigById(ctx, id)
		if err != nil {
			continue
		}
		if m.ConfigType == "Y" { // 对位 config_type=Y 内置参数不可删（Java configServiceImpl 检查）
			return errf("内置参数'%s'不能删除", m.ConfigName)
		}
		keys[m.ConfigKey] = true
	}
	if err := s.config.DeleteConfigByIds(ctx, ids); err != nil {
		return err
	}
	for k := range keys {
		_, _ = s.cache.Delete(ctx, constant.SysConfigKey+k)
	}
	return nil
}

// RefreshCache 清除全部参数缓存（对位 DELETE /refreshCache）。
func (s *ConfigService) RefreshCache(ctx context.Context) error {
	keys, err := s.cache.KeysByPrefix(ctx, constant.SysConfigKey)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		_, err = s.cache.Delete(ctx, keys...)
	}
	return err
}

// strings 引用守护。
var _ = strings.TrimSpace
