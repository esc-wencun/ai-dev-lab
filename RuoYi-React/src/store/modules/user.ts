import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { login as loginApi, getInfo as getInfoApi, logout as logoutApi } from '@/api/login'
import { getToken, setToken as persistToken, removeToken } from '@/utils/auth'
import cache from '@/plugins/cache'
import defaultAvatar from '@/assets/images/profile.jpg'

// user slice —— 对位 RuoYi-Vue3 src/store/modules/user.js（行为逐条对齐）

export interface UserState {
  token: string | undefined
  id: number | string
  name: string
  nickName: string
  avatar: string
  roles: string[]
  permissions: string[]
}

const initialState: UserState = {
  token: getToken(),
  id: 0,
  name: '',
  nickName: '',
  avatar: '',
  roles: [],
  permissions: [],
}

export const login = createAsyncThunk(
  'user/login',
  async ({ username, password, code, uuid }: { username: string; password: string; code?: string; uuid?: string }) => {
    const res = (await loginApi(username.trim(), password, code, uuid)) as unknown as { token: string }
    const token = res.token
    persistToken(token)
    return token
  },
)

export const getInfo = createAsyncThunk('user/getInfo', async () => {
  const res = (await getInfoApi()) as unknown as {
    user: { userId: number | string; userName: string; nickName: string; avatar?: string }
    roles: string[]
    permissions: string[]
    pwdChrtype?: string
    isDefaultModifyPwd?: boolean
    isPasswordExpired?: boolean
  }
  const user = res.user
  // avatar 三分支：http(s) 直用 / 空→默认图 / 否则拼 VITE_APP_BASE_API 前缀（与基准一致）
  let avatar = user.avatar || ''
  if (avatar && !/^(https?:)/.test(avatar)) {
    avatar = import.meta.env.VITE_APP_BASE_API + avatar
  }
  if (!avatar) {
    avatar = defaultAvatar
  }
  // roles 空数组 → ['ROLE_DEFAULT']（permissions 不同步赋值——与基准一致）
  let roles: string[]
  let permissions: string[]
  if (Array.isArray(res.roles) && res.roles.length > 0) {
    roles = res.roles
    permissions = res.permissions || []
  } else {
    roles = ['ROLE_DEFAULT']
    permissions = []
  }
  if (res.pwdChrtype !== undefined && res.pwdChrtype !== null) {
    // 基准键名即如此拼写（pwr 非 pwd）
    cache.session.set('pwrChrtype', String(res.pwdChrtype))
  }
  return {
    id: user.userId,
    name: user.userName,
    nickName: user.nickName,
    avatar,
    roles,
    permissions,
    isDefaultModifyPwd: !!res.isDefaultModifyPwd,
    isPasswordExpired: !!res.isPasswordExpired,
  }
})

export const logOut = createAsyncThunk('user/logOut', async () => {
  await logoutApi()
  removeToken()
})

const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    clearUser(state) {
      state.token = undefined
      state.roles = []
      state.permissions = []
      removeToken()
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.fulfilled, (state, action) => {
        state.token = action.payload
      })
      .addCase(getInfo.fulfilled, (state, action) => {
        const { isDefaultModifyPwd: _d, isPasswordExpired: _p, ...user } = action.payload
        void _d
        void _p
        Object.assign(state, user)
      })
      .addCase(logOut.fulfilled, (state) => {
        state.token = undefined
        state.roles = []
        state.permissions = []
      })
  },
})

export const { clearUser } = userSlice.actions
export default userSlice.reducer
