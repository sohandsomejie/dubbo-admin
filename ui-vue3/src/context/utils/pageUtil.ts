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

import context from '@/context'

export const getRowData = <T = Record<string, unknown>>(key?: string): T | undefined => {
  if (key) {
    return context.page.rowData[key] as T | undefined
  }
  return undefined
}

export const setRowData = <T = Record<string, unknown>>(key: string, data: T): void => {
  context.page.rowData[key] = data as Record<string, unknown>
}

export const updateRowData = (key: string, updates: Record<string, unknown>): void => {
  const current = context.page.rowData[key] || {}
  context.page.rowData[key] = { ...current, ...updates }
}

export const getRowDataKey = <T = unknown>(rowKey: string, dataKey: string): T | undefined => {
  const rowData = context.page.rowData[rowKey] as Record<string, unknown> | undefined
  return rowData?.[dataKey] as T | undefined
}

export const setRowDataKey = <T = unknown>(
  rowKey: string,
  dataKey: string,
  value: T
): void => {
  if (!context.page.rowData[rowKey]) {
    context.page.rowData[rowKey] = {}
  }
  context.page.rowData[rowKey][dataKey] = value as unknown
}

export const getInfo = (key?: string): string | undefined => {
  if (key) {
    return context.page.info[key]
  }
  return undefined
}

export const setInfo = (key: string, info: string): void => {
  context.page.info[key] = info
}

export const setInfoKey = (infoKey: string, dataKey: string, value: string): void => {
  if (!context.page.info[infoKey]) {
    context.page.info[infoKey] = ''
  }
  context.page.info[infoKey] += `\n${dataKey}: ${value}`
}

export const getInfoKey = (infoKey: string, dataKey: string): string | undefined => {
  const info = context.page.info[infoKey]
  if (!info) return undefined
  const regex = new RegExp(`${dataKey}:\\s*(.+)`)
  const match = info.match(regex)
  return match ? match[1] : undefined
}

export const clearPageContext = (): void => {
  context.page.rowData = {}
  context.page.info = {}
}

export const buildPageInfoFromRowData = (
  rowKey: string,
  infoKey: string,
  config?: {
    excludeKeys?: string[]
    includeKeys?: string[]
    transformFn?: (key: string, value: unknown) => string | null
  }
): void => {
  const { excludeKeys = [], includeKeys, transformFn } = config || {}
  const data = context.page.rowData[rowKey] as Record<string, unknown>

  if (!data) return

  let entries = Object.entries(data)

  if (includeKeys && includeKeys.length > 0) {
    entries = entries.filter(([key]) => includeKeys.includes(key))
  }

  if (excludeKeys.length > 0) {
    entries = entries.filter(([key]) => !excludeKeys.includes(key))
  }

  const lines: string[] = []

  for (const [key, value] of entries) {
    if (transformFn) {
      const transformed = transformFn(key, value)
      if (transformed !== null) {
        lines.push(`${key}: ${transformed}`)
      }
    } else {
      if (Array.isArray(value)) {
        lines.push(`${key}: ${value.length > 0 ? value.join(', ') : '(empty)'}`)
      } else if (typeof value === 'object' && value !== null) {
        lines.push(`${key}: ${JSON.stringify(value)}`)
      } else {
        lines.push(`${key}: ${value ?? '(empty)'}`)
      }
    }
  }

  context.page.info[infoKey] = lines.join('\n')
}
