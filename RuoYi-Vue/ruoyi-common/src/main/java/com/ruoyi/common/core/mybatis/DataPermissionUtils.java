package com.ruoyi.common.core.mybatis;

import java.util.concurrent.Callable;
import com.ruoyi.common.core.domain.dto.RuoYiDeptDataPermissionDTO;
import com.ruoyi.common.exception.UtilException;

/**
 * 数据权限编程式豁免工具：作用域内拦截器不注入任何范围条件。
 * 典型场景：唯一性校验、系统内部数据修补等需要跨范围读数据的逻辑。
 *
 * @author ruoyi
 */
public class DataPermissionUtils
{
    private DataPermissionUtils()
    {
    }

    /** 作用域内豁免数据权限（无返回值） */
    public static void executeIgnore(Runnable runnable)
    {
        executeIgnore(() -> {
            runnable.run();
            return null;
        });
    }

    /** 作用域内豁免数据权限（有返回值） */
    public static <T> T executeIgnore(Callable<T> callable)
    {
        DataScopeContextHolder.add(new DataScopeContextHolder.Scope(false, null));
        try
        {
            return callable.call();
        }
        catch (RuntimeException e)
        {
            throw e;
        }
        catch (Exception e)
        {
            throw new UtilException(e);
        }
        finally
        {
            DataScopeContextHolder.remove();
        }
    }
}
