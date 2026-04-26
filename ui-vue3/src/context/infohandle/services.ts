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
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express in writing
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { setRowData, setInfo } from '@/context/utils/pageUtil'

export interface ServiceSearchItem {
  serviceName: string
  versionGroups: Array<{ version: string | null; group: string | null }>
  avgQPS: number
  avgRT: string
  requestTotal: number
}

export interface ServiceDistributionItem {
  appName: string
  instanceCount: number
  instanceName: string
  rpcPort: string
  timeout: string
  retryNum: string
  label: string
}

export const generateServicesListInfo = (data: ServiceSearchItem[]): string => {
  const lines: string[] = []
  
  lines.push(`Services List (${data.length} total):`)
  lines.push('')
  
  data.forEach((item, index) => {
    lines.push(`[${index + 1}] Service: ${item.serviceName}`)
    lines.push(`    Versions: ${item.versionGroups?.map(vg => `${vg.version || 'N/A'}/${vg.group || 'N/A'}`).join(', ') || 'N/A'}`)
    lines.push(`    Avg QPS: ${item.avgQPS || 0}`)
    lines.push(`    Avg RT: ${item.avgRT || 'N/A'}`)
    lines.push(`    Total Requests: ${item.requestTotal || 0}`)
    lines.push('')
  })

  return lines.join('\n')
}

export const buildServicesListInfo = (data: ServiceSearchItem[]): void => {
  setRowData('list', data)
  setInfo('list', generateServicesListInfo(data))
}

export const generateServiceDistributionInfo = (data: ServiceDistributionItem[]): string => {
  const lines: string[] = []
  
  lines.push(`Service Distribution (${data.length} entries):`)
  lines.push('')
  
  data.forEach((item, index) => {
    lines.push(`[${index + 1}] Application: ${item.appName}`)
    lines.push(`    Instance: ${item.instanceName} (${item.instanceCount} instances)`)
    lines.push(`    RPC Port: ${item.rpcPort || 'N/A'}`)
    lines.push(`    Timeout: ${item.timeout || 'N/A'}`)
    lines.push(`    Retry: ${item.retryNum || 'N/A'}`)
    lines.push(`    Cluster: ${item.label || 'N/A'}`)
    lines.push('')
  })

  return lines.join('\n')
}

export const buildServiceDistributionInfo = (data: ServiceDistributionItem[]): void => {
  setRowData('distribution', data)
  setInfo('distribution', generateServiceDistributionInfo(data))
}

export const generateServiceDebugInfo = (
  serviceName: string,
  debugData: Record<string, unknown>
): string => {
  const lines: string[] = []

  lines.push(`Service Debug: ${serviceName}`)
  lines.push('')
  
  if (debugData.methods) {
    const methods = Array.isArray(debugData.methods) ? debugData.methods : []
    lines.push(`Available Methods (${methods.length}):`)
    methods.forEach((method: any) => {
      lines.push(`  - ${method.methodName || method}`)
    })
    lines.push('')
  }

  if (debugData.instances) {
    const instances = Array.isArray(debugData.instances) ? debugData.instances : []
    lines.push(`Provider Instances (${instances.length}):`)
    instances.forEach((instance: any) => {
      lines.push(`  - ${instance.name || instance.ip || instance}`)
    })
  }

  return lines.join('\n')
}

export const buildServiceDebugInfo = (
  serviceName: string,
  debugData: Record<string, unknown>
): void => {
  setRowData('debug', { serviceName, ...debugData })
  setInfo('debug', generateServiceDebugInfo(serviceName, debugData))
}

export const generateServiceTopologyInfo = (
  serviceName: string,
  graphData: Record<string, unknown>
): string => {
  const lines: string[] = []

  lines.push(`Service Topology: ${serviceName}`)
  lines.push('')
  
  if (graphData.nodes) {
    const nodes = Array.isArray(graphData.nodes) ? graphData.nodes : []
    lines.push(`Total Nodes: ${nodes.length}`)
  }

  if (graphData.edges) {
    const edges = Array.isArray(graphData.edges) ? graphData.edges : []
    lines.push(`Total Connections: ${edges.length}`)
  }

  return lines.join('\n')
}

export const buildServiceTopologyInfo = (
  serviceName: string,
  graphData: Record<string, unknown>
): void => {
  setRowData('topology', { serviceName, ...graphData })
  setInfo('topology', generateServiceTopologyInfo(serviceName, graphData))
}

export const generateServiceMonitorInfo = (serviceName: string): string => {
  const lines: string[] = []
  lines.push(`Monitoring Dashboard for Service: ${serviceName}`)
  lines.push('')
  lines.push(`Metrics data is loaded from Grafana`)
  return lines.join('\n')
}

export const buildServiceMonitorInfo = (serviceName: string): void => {
  setRowData('monitor', { serviceName })
  setInfo('monitor', generateServiceMonitorInfo(serviceName))
}

export const generateServiceTracingInfo = (serviceName: string): string => {
  const lines: string[] = []
  lines.push(`Distributed Tracing for Service: ${serviceName}`)
  lines.push('')
  lines.push(`Trace data is loaded from Grafana`)
  return lines.join('\n')
}

export const buildServiceTracingInfo = (serviceName: string): void => {
  setRowData('tracing', { serviceName })
  setInfo('tracing', generateServiceTracingInfo(serviceName))
}

export const generateServiceSceneConfigInfo = (config: Record<string, unknown>): string => {
  const lines: string[] = []

  lines.push(`Scene Configuration for Service: ${config.serviceName || 'Unknown'}`)
  lines.push('')

  if (config.timeout !== undefined) {
    lines.push(`Timeout: ${config.timeout}ms`)
  }

  if (config.retry !== undefined) {
    lines.push(`Retry Count: ${config.retry}`)
  }

  if (config.regionPriority !== undefined) {
    lines.push(`Intra-region Priority: ${config.regionPriority ? 'Enabled' : 'Disabled'}`)
  }

  if (config.routes && Array.isArray(config.routes)) {
    lines.push('')
    lines.push(`Argument Routes (${config.routes.length} configured):`)
    config.routes.forEach((route: any, index: number) => {
      lines.push(`  Route ${index + 1}:`)
      if (route.method) {
        lines.push(`    Method: ${route.method}`)
      }
      if (route.conditions) {
        lines.push(`    Conditions: ${JSON.stringify(route.conditions)}`)
      }
      if (route.destinations) {
        route.destinations.forEach((dest: any, destIndex: number) => {
          lines.push(`    Destination ${destIndex + 1}:`)
          lines.push(`      Weight: ${dest.weight}`)
          lines.push(`      Match: ${JSON.stringify(dest.conditions)}`)
        })
      }
    })
  }

  return lines.join('\n')
}

export const buildServiceSceneConfigInfo = (config: Record<string, unknown>): void => {
  setRowData('sceneConfig', config)
  setInfo('sceneConfig', generateServiceSceneConfigInfo(config))
}
