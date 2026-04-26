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

import { reactive } from 'vue'
import type { AppContext, GlobalContext, PageContext, RouteContext } from './types'

const initialContext: AppContext = {
  global: {},
  page: {
    rowData: {},
    info: {}
  },
  route: {
    path: '',
    fullPath: '',
    params: {},
    query: {}
  }
}

const context = reactive<AppContext>(initialContext)

export const getContext = (): AppContext => context

export const getGlobalContext = (): GlobalContext => context.global

export const getPageContext = (): PageContext => context.page

export const getRouteContext = (): RouteContext => context.route

export const setGlobalContext = <K extends keyof GlobalContext>(
  key: K,
  value: GlobalContext[K]
): void => {
  context.global[key] = value
}

export const setPageContext = (): void => {
  // rowData 和 info 已改为键控存储，不再使用此方法
}

export const setRouteContext = (value: Partial<RouteContext>): void => {
  context.route = { ...context.route, ...value }
}

export const resetContext = (): void => {
  context.global = {}
  context.page = { rowData: {}, info: {} }
  context.route = {
    path: '',
    fullPath: '',
    params: {},
    query: {}
  }
}

export const resetPageContext = (): void => {
  context.page = { rowData: {}, info: {} }
}

export const resetGlobalContext = (): void => {
  context.global = {}
}

export const useRouteContext = () => context.route

export default context
