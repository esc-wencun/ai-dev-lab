import Cookies from 'js-cookie'

// token 存取 —— 键名 Admin-Token 是与三版后端的兼容契约，不得改名；
// 会话级 Cookie（不设 expires），与 RuoYi-Vue3 src/utils/auth.js 一致
const TokenKey = 'Admin-Token'

export function getToken() {
  return Cookies.get(TokenKey)
}

export function setToken(token: string) {
  return Cookies.set(TokenKey, token)
}

export function removeToken() {
  return Cookies.remove(TokenKey)
}
