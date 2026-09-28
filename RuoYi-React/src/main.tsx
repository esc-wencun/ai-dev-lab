import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router'
import { App as AntdApp } from 'antd'
import '@ant-design/v5-patch-for-react-19'
import 'virtual:svg-icons-register'

import App from './App.tsx'
import { store } from './store'
import ThemeBridge from './components/ThemeBridge'

// 调试用：浏览器控制台可读 store（生产构建自动剔除）
if (import.meta.env.DEV) {
  ;(window as unknown as { __ry_store: unknown }).__ry_store = store
}
import './assets/styles/index.scss'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Provider store={store}>
      <BrowserRouter>
        {/* ThemeBridge：Provider 内读 settings，动态装配 ConfigProvider（明暗算法 + 主题色 + 中文 locale） */}
        <ThemeBridge>
          <AntdApp>
            <App />
          </AntdApp>
        </ThemeBridge>
      </BrowserRouter>
    </Provider>
  </StrictMode>,
)
