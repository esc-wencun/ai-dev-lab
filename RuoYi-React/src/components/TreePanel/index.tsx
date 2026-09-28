// TreePanel —— 对位基准 components/TreePanel/index.vue（可折叠、可搜索的树侧栏）
// 基准唯一消费页是 user 页左树：页面经 deptTreeSelect 拉取部门树并以 treeData 传入（组件不自取接口，读基准源码确认）
// props/events 对位：treeData/title/searchPlaceholder/defaultExpandAll/node-click/refresh/expose.setCurrentKey
// 差异（对齐 deviations #4 纯视觉糖简化）：
//   1. 基准支持拖拽调宽 + localStorage 宽度持久化——纯布局糖，不实现（宽度固定）；
//   2. 树节点 Folder/Document 图标省略，仅保留文本标签；
//   3. 侧栏整体折叠按钮（collapse-button）省略——唯一消费页未使用该能力，功能无损失
// title 默认对位基准「树形结构」，user 页传入「组织机构」
// expose 等价：经 ref 暴露 setCurrentKey/clearSearch（基准 user 页 resetQuery 消费 setCurrentKey(null)）

import { forwardRef, useEffect, useImperativeHandle, useState } from 'react'
import { Card, Input, Tree } from 'antd'
import { SearchOutlined, ArrowUpOutlined, ArrowDownOutlined, SyncOutlined } from '@ant-design/icons'
import type { DataNode } from 'antd/es/tree'

export interface TreePanelRef {
  /** 对位基准 expose setCurrentKey：null 清除高亮 */
  setCurrentKey: (key: React.Key | null) => void
  /** 对位基准 expose clearSearch：清空搜索框并取消过滤 */
  clearSearch: () => void
}

interface TreePanelProps {
  /** 树形数据（对位 treeData prop） */
  treeData: DataNode[]
  /** 标题，默认「树形结构」 */
  title?: string
  /** 搜索框占位符 */
  searchPlaceholder?: string
  /** 是否默认展开全部，默认 false */
  defaultExpandAll?: boolean
  /** 节点点击回调（对位 @node-click，参数为节点 key） */
  onNodeClick?: (key: React.Key) => void
  /** 头部刷新按钮回调（对位 @refresh） */
  onRefresh?: () => void
}

// 从树数据递归收集全部节点 key（展开全部/收起全部用）
function collectKeys(items: DataNode[], acc: React.Key[] = []): React.Key[] {
  items.forEach((n) => {
    acc.push(n.key)
    if (n.children) collectKeys(n.children as DataNode[], acc)
  })
  return acc
}

export default forwardRef<TreePanelRef, TreePanelProps>(function TreePanel(
  { treeData, title = '树形结构', searchPlaceholder, defaultExpandAll = false, onNodeClick, onRefresh }: TreePanelProps,
  ref,
) {
  const [searchKeyword, setSearchKeyword] = useState('')
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([])
  // 全部展开/收起状态（对位基准 isExpandedAll，默认跟随 defaultExpandAll）
  const [expandAll, setExpandAll] = useState(defaultExpandAll)

  // 受控展开：跟随 expandAll 切换；树数据变化（刷新/首次加载完成）时同步一次
  useEffect(() => {
    setExpandedKeys(expandAll ? collectKeys(treeData) : [])
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [treeData])

  // 搜索过滤（对位基准 el-tree filter-node-method：按 label 包含匹配）
  const filteredData = (() => {
    if (!searchKeyword) return treeData
    const match = (items: DataNode[]): DataNode[] => {
      const out: DataNode[] = []
      for (const n of items) {
        const children = n.children ? match(n.children as DataNode[]) : undefined
        const label = typeof n.title === 'string' ? n.title : ''
        if ((label && label.indexOf(searchKeyword) !== -1) || (children && children.length > 0)) {
          out.push({ ...n, children })
        }
      }
      return out
    }
    return match(treeData)
  })()

  // 搜索时自动展开命中链路，保证过滤结果可见
  useEffect(() => {
    if (searchKeyword) {
      setExpandedKeys(collectKeys(filteredData))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchKeyword])

  useImperativeHandle(ref, () => ({
    setCurrentKey: (key) => {
      // setSelectedKeys 空数组即清除高亮（antd Tree 无 setCurrentKey，语义等价）
      setSelKeys(key === null ? [] : [key])
    },
    clearSearch: () => {
      setSearchKeyword('')
    },
  }))

  // 选中高亮（对位 el-tree current-node 高亮 + setCurrentKey）
  const [selKeys, setSelKeys] = useState<React.Key[]>([])

  const toggleExpandAll = () => {
    const next = !expandAll
    setExpandAll(next)
    setExpandedKeys(next ? collectKeys(treeData) : [])
  }

  return (
    <Card
      size="small"
      title={title}
      extra={
        <span>
          {/* 展开全部/收起全部图标（对位基准：展开中显示 ArrowDown，收起状态显示 ArrowUp） */}
          {expandAll ? (
            <ArrowDownOutlined
              title="收起全部"
              onClick={toggleExpandAll}
              style={{ cursor: 'pointer', color: '#909399', marginRight: 8 }}
            />
          ) : (
            <ArrowUpOutlined
              title="展开全部"
              onClick={toggleExpandAll}
              style={{ cursor: 'pointer', color: '#909399', marginRight: 8 }}
            />
          )}
          <SyncOutlined title="刷新" onClick={() => onRefresh?.()} style={{ cursor: 'pointer', color: '#909399' }} />
        </span>
      }
    >
      <Input
        allowClear
        prefix={<SearchOutlined style={{ color: '#c0c4cc' }} />}
        placeholder={searchPlaceholder}
        value={searchKeyword}
        onChange={(e) => setSearchKeyword(e.target.value)}
        style={{ marginBottom: 8 }}
      />
      <div style={{ overflowY: 'auto' }}>
        <Tree
          treeData={filteredData}
          showLine={false}
          expandedKeys={expandedKeys}
          onExpand={(keys) => setExpandedKeys(keys)}
          selectedKeys={selKeys}
          onSelect={(keys) => {
            // el-tree 点已选中节点仍触发 node-click；antd onSelect 取消选中时返回空数组，保持高亮不改 deptId
            if (keys.length > 0 && onNodeClick) onNodeClick(keys[keys.length - 1])
          }}
        />
      </div>
    </Card>
  )
})
