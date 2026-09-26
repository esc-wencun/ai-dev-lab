package com.ruoyi.system.mapper;

import java.util.List;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.ruoyi.common.core.mybatis.LambdaQueryWrapperX;
import com.ruoyi.common.utils.PageUtils;
import com.ruoyi.system.domain.SysPost;

/**
 * 岗位信息 数据层
 *
 * @author ruoyi
 */
public interface SysPostMapper extends BaseMapper<SysPost>
{
    /**
     * 分页查询岗位列表（动态条件用 Wrapper 组装，替代 XML selectPostList 的分页场景）
     *
     * @param post 岗位信息（postCode/postName 模糊、status 精确，值空自动跳过）
     * @return 分页结果
     */
    default Page<SysPost> selectPostPage(SysPost post)
    {
        return selectPage(PageUtils.buildPage(), new LambdaQueryWrapperX<SysPost>()
                .likeIfPresent(SysPost::getPostCode, post.getPostCode())
                .likeIfPresent(SysPost::getPostName, post.getPostName())
                .eqIfPresent(SysPost::getStatus, post.getStatus())
                .orderByAsc(SysPost::getPostSort));
    }

    /**
     * 查询岗位数据集合（非分页场景：导出等）
     *
     * @param post 岗位信息
     * @return 岗位数据集合
     */
    List<SysPost> selectPostList(SysPost post);

    /**
     * 查询所有岗位
     *
     * @return 岗位列表
     */
    List<SysPost> selectPostAll();

    /**
     * 根据用户ID获取岗位选择框列表
     *
     * @param userId 用户ID
     * @return 选中岗位ID列表
     */
    List<Long> selectPostListByUserId(Long userId);

    /**
     * 查询用户所属岗位组
     *
     * @param userName 用户名
     * @return 结果
     */
    List<SysPost> selectPostsByUserName(String userName);
}
