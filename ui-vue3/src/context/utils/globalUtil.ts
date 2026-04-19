import { LOCAL_STORAGE_LOCALE } from '@/base/constants'
import Cookies from 'js-cookie'
import context from '@/context'

export const syncAuthContext = () => {
  const authState = JSON.parse(Cookies.get('auth-state') || '{}')
  const username = authState?.userinfo?.username
  if (username) {
    context.global.user = { username: String(username) }
  } else {
    delete context.global.user
  }
}

export const syncI18nContext = () => {
  const locale = localStorage.getItem(LOCAL_STORAGE_LOCALE) || 'cn'
  context.global.i18n = { locale }
}

