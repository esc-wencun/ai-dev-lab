// 首页 —— 对位基准 index.vue 的静态介绍页（文案归属本工程，deviations #10）
// 布局功能等价：版本卡 + 技术选型 + 更新日志折叠；捐赠卡去除

import { Card, Col, Collapse, Row, Space, Tag, Typography } from 'antd'

const version = '1.0.0'

export default function Index() {
  return (
    <div style={{ padding: 24 }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card>
            <Space direction="vertical" size="small">
              <Typography.Title level={4} style={{ margin: 0 }}>
                RuoYi-React 后台管理系统 <Tag color="primary">v{version}</Tag>
              </Typography.Title>
              <Typography.Text type="secondary">
                React 19 + TypeScript + Ant Design 5 + Redux Toolkit 实现的若依管理前端，
                与 RuoYi-Vue3 功能等价，服务 Java / Python / Go 三版后端。
              </Typography.Text>
            </Space>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title="技术选型">
            <Space wrap>
              {['React 19', 'TypeScript', 'Ant Design 5', 'Redux Toolkit', 'React Router 7', 'Axios', 'Vite', 'Vitest', 'ECharts'].map(
                (t) => (
                  <Tag key={t}>{t}</Tag>
                ),
              )}
            </Space>
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card title="更新日志">
            <Collapse
              items={[
                {
                  key: '1',
                  label: `v${version}（2026-09-27）`,
                  children: (
                    <ol>
                      <li>工程基础：Vite + React 19 + TS 脚手架、svg 精灵同源、8090 端口</li>
                      <li>基础设施：请求层契约逐条复刻（401 确认框/防重/下载）、25 项单测</li>
                      <li>状态与路由：RTK 七 store、动态路由 Gate 方案、守卫与锁屏劫持</li>
                      <li>登录闭环：验证码 / 记住我 RSA / redirect 跳转</li>
                    </ol>
                  ),
                },
              ]}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
