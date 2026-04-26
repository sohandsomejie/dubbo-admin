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

import { setRowData, setInfo } from '@/context/utils/pageUtil'

export interface HomeOverviewData {
  appCount?: number
  serviceCount?: number
  instanceCount?: number
  protocols?: string[]
  releases?: string[]
  discoveries?: string[]
}

export interface HomeMetadataData {
  registry?: string
  metadata?: string
  config?: string
  prometheus?: string
  grafana?: string
  tracing?: string
}

export const generateHomeOverviewInfo = (data: HomeOverviewData): string => {
  const lines: string[] = []

  lines.push(`Cluster Overview:`)
  lines.push(`  Total Applications: ${data.appCount ?? 'N/A'} (registered Dubbo applications)`)
  lines.push(`  Total Services: ${data.serviceCount ?? 'N/A'} (Dubbo service interfaces)`)
  lines.push(`  Total Instances: ${data.instanceCount ?? 'N/A'} (running service providers/consumers)`)
  lines.push('')
  
  if (data.protocols && data.protocols.length > 0) {
    lines.push(`Protocols: ${data.protocols.join(', ')} (e.g., triple, dubbo, gRPC)`)
  }
  
  if (data.releases && data.releases.length > 0) {
    lines.push(`Dubbo Versions: ${data.releases.join(', ')}`)
  }
  
  if (data.discoveries && data.discoveries.length > 0) {
    lines.push(`Registry Centers: ${data.discoveries.join(', ')} (service discovery endpoints)`)
  }

  return lines.join('\n')
}

export const buildHomeOverviewInfo = (data: HomeOverviewData): void => {
  setRowData('overview', data)
  setInfo('overview', generateHomeOverviewInfo(data))
}

export const generateHomeMetadataInfo = (data: HomeMetadataData): string => {
  const lines: string[] = []

  lines.push(`Metadata Configuration:`)
  lines.push('')
  
  if (data.registry) {
    lines.push(`Service Registry: ${data.registry} (service registration and discovery center)`)
  }
  
  if (data.metadata) {
    lines.push(`Metadata Center: ${data.metadata} (service metadata storage)`)
  }
  
  if (data.config) {
    lines.push(`Config Center: ${data.config} (dynamic configuration management)`)
  }
  
  if (data.prometheus) {
    lines.push(`Prometheus: ${data.prometheus} (metrics collection endpoint)`)
  }
  
  if (data.grafana) {
    lines.push(`Grafana: ${data.grafana} (metrics visualization dashboard)`)
  }
  
  if (data.tracing) {
    lines.push(`Distributed Tracing: ${data.tracing} (link tracking and analysis)`)
  }

  return lines.join('\n')
}

export const buildHomeMetadataInfo = (data: HomeMetadataData): void => {
  setRowData('metadata', data)
  setInfo('metadata', generateHomeMetadataInfo(data))
}
