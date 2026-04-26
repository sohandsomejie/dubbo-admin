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

export interface GlobalContext {
  user?: {
    username: string
  }
  i18n?: {
    locale: 'en' | 'cn'
  }
}

export interface PageContext {
  rowData: Record<string, Record<string, unknown>>
  info: Record<string, string>
}

export interface RouteMeta {
  icon?: string
  hidden?: boolean
  skip?: boolean
  tab_parent?: boolean
  tab?: boolean
  back?: string
  headerParamKey?: string
  [key: string]: unknown
}

export interface RouteContext {
  name?: string
  path: string
  fullPath: string
  params: Record<string, string>
  query: Record<string, string>
  meta?: RouteMeta
}

export interface AppContext {
  global: GlobalContext
  page: PageContext
  route: RouteContext
}
