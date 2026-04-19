import type { RouteLocationNormalized } from 'vue-router'
import context from '@/context'

export const pickRouteMeta = (meta: Record<string, unknown>) => {
  const allowedKeys = [
    'icon',
    'hidden',
    'skip',
    'tab_parent',
    'tab',
    '_router_key',
    'headerParamKey',
    'back'
  ] as const

  const safeMeta: Record<string, unknown> = {}
  for (const key of allowedKeys) {
    const value = meta[key]
    if (value === undefined) {
      continue
    }
    if (
      value === null ||
      typeof value === 'string' ||
      typeof value === 'number' ||
      typeof value === 'boolean' ||
      Array.isArray(value)
    ) {
      safeMeta[key] = value
      continue
    }
    safeMeta[key] = String(value)
  }
  return safeMeta
}

export const syncRouteContext = (to: RouteLocationNormalized) => {
  context.route = {
    name: to.name ? String(to.name) : undefined,
    path: to.path,
    fullPath: to.fullPath,
    params: to.params,
    query: to.query,
    meta: pickRouteMeta(to.meta as Record<string, unknown>)
  }
  console.log(context)
}
