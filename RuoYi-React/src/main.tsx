import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router'
import { ConfigProvider, App as AntdApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import '@ant-design/v5-patch-for-react-19'
import 'virtual:svg-icons-register'

import App from './App.tsx'
import { store } from './store'

// 调试用：浏览器控制台可读 store（生产构建自动剔除）
if (import.meta.env.DEV) {
  ;(window as unknown as { __ry_store: unknown }).__ry_store = store
}
import './assets/styles/index.scss'

const theme = {
  token: {
    colorPrimary: '#409EFF',
  },
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Provider store={store}>
      <BrowserRouter>
        <ConfigProvider locale={zhCN} theme={theme}>
          <AntdApp>
            <App />
          </AntdApp>
        </ConfigProvider>
      </BrowserRouter>
    </Provider>
  </StrictMode>,
)
