/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import type { RouteLocationNormalized } from 'vue-router'
import context, { setRouteContext, useRouteContext } from '@/context'
import type { RouteMeta } from '@/context/types'

const ALLOWED_META_KEYS: (keyof RouteMeta)[] = [
  'icon',
  'hidden',
  'skip',
  'tab_parent',
  'tab',
  'back',
  'headerParamKey'
]

export const pickRouteMeta = (meta: Record<string, unknown>): RouteMeta => {
  const safeMeta: RouteMeta = {}

  for (const key of ALLOWED_META_KEYS) {
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
      safeMeta[key] = value as RouteMeta[typeof key]
    } else {
      safeMeta[key] = String(value)
    }
  }

  return safeMeta
}

export const syncRouteContext = (to: RouteLocationNormalized): void => {
  setRouteContext({
    name: to.name ? String(to.name) : undefined,
    path: to.path,
    fullPath: to.fullPath,
    params: to.params as Record<string, string>,
    query: to.query as Record<string, string>,
    meta: pickRouteMeta(to.meta as Record<string, unknown>)
  })
  console.log(context)
}

export const getRouteName = (): string | undefined => {
  return useRouteContext().name
}

export const getRoutePath = (): string => {
  return useRouteContext().path
}

export const getRouteParams = <T = Record<string, string>>(): T => {
  return useRouteContext().params as T
}

export const getRouteQuery = <T = Record<string, string>>(): T => {
  return useRouteContext().query as T
}

export const getRouteMeta = (): RouteMeta | undefined => {
  return useRouteContext().meta
}

export { context }
