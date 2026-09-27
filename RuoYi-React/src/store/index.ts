import { configureStore } from '@reduxjs/toolkit'
import dict from './modules/dict'
import user from './modules/user'
import permission from './modules/permission'
import app from './modules/app'
import settings from './modules/settings'
import tagsView from './modules/tagsView'
import lock from './modules/lock'

export const store = configureStore({
  reducer: {
    dict,
    user,
    permission,
    app,
    settings,
    tagsView,
    lock,
  },
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
