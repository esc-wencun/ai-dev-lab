package com.ruoyi.framework.aspectj;

import java.util.HashSet;
import java.util.Set;
import java.util.stream.Collectors;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import com.ruoyi.common.constant.Constants;
import com.ruoyi.common.constant.UserConstants;
import com.ruoyi.common.core.domain.dto.RuoYiDeptDataPermissionDTO;
import com.ruoyi.common.core.domain.entity.SysRole;
import com.ruoyi.common.core.domain.entity.SysUser;
import com.ruoyi.common.core.domain.model.LoginUser;
import com.ruoyi.common.core.text.Convert;
import com.ruoyi.common.utils.SecurityUtils;
import com.ruoyi.common.utils.StringUtils;
import com.ruoyi.common.core.domain.entity.SysDept;
import com.ruoyi.system.service.ISysDeptService;

/**
 * 部门数据权限范围解析器：把当前登录用户的角色范围解析为结构化 DTO
 * （{all, self, deptIds}），供 MP DataPermissionInterceptor 逐表生成过滤条件。
 *
 * <p>角色判断结构与 {@link DataScopeAspect} 保持一致（admin 旁路、停用角色跳过、
 * permission 串过滤、多角色并集、任一 ALL 即 ALL）；角色数据源同为
 * LoginUser.user.getRoles() 登录快照，保证双路径行为一致。
 *
 * <p>与切面 SQL 片段的实现差异（结果语义一致，端到端专测覆盖）：
 * 自定义范围按 sys_role_dept 展开 ID 集、本部门及以下按 selectChildrenDeptById 预展开，均用 IN 而非子查询/find_in_set。
 *
 * @author ruoyi
 */
@Component
public class DeptDataScopeResolver
{
    @Autowired
    private ISysDeptService deptService;

    /**
     * 解析当前登录用户的数据权限范围。
     *
     * @param permission 权限字符（空则不过滤角色；非空则仅计匹配角色，与切面语义一致）
     * @return 范围 DTO；非 Web 上下文（无登录态）返回 null，调用方应整体放行
     */
    public RuoYiDeptDataPermissionDTO resolve(String permission)
    {
        LoginUser loginUser;
        try
        {
            loginUser = SecurityUtils.getLoginUser();
        }
        catch (Exception e)
        {
            return null;   // 非 Web 上下文（定时任务等）放行
        }
        if (loginUser == null)
        {
            return null;
        }
        SysUser user = loginUser.getUser();
        if (user == null)
        {
            return null;
        }
        // 超级管理员不过滤
        if (user.isAdmin())
        {
            return RuoYiDeptDataPermissionDTO.all();
        }

        boolean all = false;
        boolean self = false;
        Set<Long> deptIds = new HashSet<>();
        boolean anyRoleMatched = false;

        for (SysRole role : user.getRoles())
        {
            if (StringUtils.equals(role.getStatus(), UserConstants.ROLE_DISABLE))
            {
                continue;
            }
            if (StringUtils.isNotEmpty(permission) && !StringUtils.containsAny(role.getPermissions(), Convert.toStrArray(permission)))
            {
                continue;
            }
            anyRoleMatched = true;
            String dataScope = role.getDataScope();
            if (Constants.Dept.DATA_SCOPE_ALL.equals(dataScope))
            {
                all = true;
                break;
            }
            else if (Constants.Dept.DATA_SCOPE_CUSTOM.equals(dataScope))
            {
                // 数据权限场景取全量自定义部门集（strictly=false：含父部门——与切面 IN 子查询语义一致）
                deptIds.addAll(deptService.selectDeptListByRoleId(role.getRoleId()));
            }
            else if (Constants.Dept.DATA_SCOPE_DEPT.equals(dataScope))
            {
                deptIds.add(user.getDeptId());
            }
            else if (Constants.Dept.DATA_SCOPE_DEPT_AND_CHILD.equals(dataScope))
            {
                deptIds.add(user.getDeptId());
                deptIds.addAll(deptService.selectChildrenDeptById(user.getDeptId())
                        .stream().map(SysDept::getDeptId).collect(Collectors.toSet()));
            }
            else if (Constants.Dept.DATA_SCOPE_SELF.equals(dataScope))
            {
                self = true;
            }
        }

        // 有角色但都不含权限字符（与切面"角色都不包含传递过来的权限字符 → 查不到任何数据"语义对齐）
        if (!anyRoleMatched)
        {
            return RuoYiDeptDataPermissionDTO.none();
        }
        if (self && deptIds.isEmpty())
        {
            return new RuoYiDeptDataPermissionDTO(false, true, null);
        }
        return new RuoYiDeptDataPermissionDTO(all, self, deptIds);
    }
}
