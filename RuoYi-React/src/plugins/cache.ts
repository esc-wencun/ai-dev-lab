// session/local 两套缓存封装 —— 对位 RuoYi-Vue3 src/plugins/cache.js，行为一致
const sessionCache = {
  set(key: string, value: string) {
    if (typeof sessionStorage === 'undefined') return
    if (key != null && value != null) {
      sessionStorage.setItem(key, value)
    }
  },
  get(key: string): string | null {
    if (typeof sessionStorage === 'undefined') return null
    if (key == null) return null
    return sessionStorage.getItem(key)
  },
  setJSON(key: string, jsonValue: unknown) {
    if (jsonValue != null) {
      this.set(key, JSON.stringify(jsonValue))
    }
  },
  getJSON<T = unknown>(key: string): T | null {
    const value = this.get(key)
    if (value != null) {
      return JSON.parse(value) as T
    }
    return null
  },
  remove(key: string) {
    sessionStorage.removeItem(key)
  },
}

const localCache = {
  set(key: string, value: string) {
    if (typeof localStorage === 'undefined') return
    if (key != null && value != null) {
      localStorage.setItem(key, value)
    }
  },
  get(key: string): string | null {
    if (typeof localStorage === 'undefined') return null
    if (key == null) return null
    return localStorage.getItem(key)
  },
  setJSON(key: string, jsonValue: unknown) {
    if (jsonValue != null) {
      this.set(key, JSON.stringify(jsonValue))
    }
  },
  getJSON<T = unknown>(key: string): T | null {
    const value = this.get(key)
    if (value != null) {
      return JSON.parse(value) as T
    }
    return null
  },
  remove(key: string) {
    localStorage.removeItem(key)
  },
}

const cache = {
  /** 会话级缓存 */
  session: sessionCache,
  /** 本地缓存 */
  local: localCache,
}

export default cache
