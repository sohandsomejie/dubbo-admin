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
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either or any
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { setRowData, setInfo } from '@/context/utils/pageUtil'

export interface ApplicationSearchItem {
  appName: string
  deployClusters: string[]
  instanceCount: number
  registryClusters: string[]
}

export interface ApplicationDetail {
  appName: string
  appTypes: string[]
  deployClusters: string[]
  dubboPorts: number[]
  dubboVersions: string[]
  images: string[]
  registerClusters: string[]
  registerModes: string[]
  rpcProtocols: string[]
  serialProtocols: string[]
  workloads: string[]
}

export interface ApplicationInstanceStatistics {
  instanceTotal: number
  versionTotal: number
  cpuTotal: string
  memoryTotal: string
}

export interface ApplicationInstanceInfoItem {
  ip: string
  name: string
  deployState: string
  deployCluster: string
  registerState: string
  registerClusters: string[]
  cpu: string
  memory: string
  startTime: string
  registerTime: string
  labels: Record<string, string>
}

export const generateApplicationsListInfo = (data: ApplicationSearchItem[]): string => {
  const lines: string[] = []
  
  lines.push(`Applications List (${data.length} total):`)
  lines.push('')
  
  data.forEach((item, index) => {
    lines.push(`[${index + 1}] Application: ${item.appName}`)
    lines.push(`    Instance Count: ${item.instanceCount}`)
    lines.push(`    Deploy Clusters: ${item.deployClusters?.join(', ') || 'N/A'}`)
    lines.push(`    Registry Clusters: ${item.registryClusters?.join(', ') || 'N/A'}`)
    lines.push('')
  })

  return lines.join('\n')
}

export const buildApplicationsListInfo = (data: ApplicationSearchItem[]): void => {
  setRowData('list', data)
  setInfo('list', generateApplicationsListInfo(data))
}

export const generateApplicationDetailInfo = (data: ApplicationDetail): string => {
  const lines: string[] = []

  lines.push(`Application: ${data.appName}`)
  
  if (data.appTypes && data.appTypes.length > 0) {
    lines.push(`Application Types: ${data.appTypes.join(', ')}`)
  }

  if (data.deployClusters && data.deployClusters.length > 0) {
    lines.push(`Deploy Clusters: ${data.deployClusters.join(', ')}`)
  }

  if (data.registerClusters && data.registerClusters.length > 0) {
    lines.push(`Registry Clusters: ${data.registerClusters.join(', ')}`)
  }

  if (data.registerModes && data.registerModes.length > 0) {
    lines.push(`Registration Modes: ${data.registerModes.join(', ')}`)
  }

  if (data.rpcProtocols && data.rpcProtocols.length > 0) {
    lines.push(`RPC Protocols: ${data.rpcProtocols.join(', ')} (e.g., triple, dubbo, gRPC)`)
  }

  if (data.dubboVersions && data.dubboVersions.length > 0) {
    lines.push(`Dubbo Versions: ${data.dubboVersions.join(', ')}`)
  }

  if (data.workloads && data.workloads.length > 0) {
    lines.push(`Workloads: ${data.workloads.join(', ')}`)
  }

  return lines.join('\n')
}

export const buildApplicationDetailInfo = (data: ApplicationDetail): void => {
  setRowData('detail', data)
  setInfo('detail', generateApplicationDetailInfo(data))
}

export const generateApplicationTopologyInfo = (graphData: Record<string, unknown>): string => {
  const lines: string[] = []

  lines.push(`Application Topology for: ${graphData.appName || 'Unknown'}`)
  
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

export const buildApplicationTopologyInfo = (graphData: Record<string, unknown>): void => {
  setRowData('topology', graphData)
  setInfo('topology', generateApplicationTopologyInfo(graphData))
}

export const generateApplicationInstanceInfo = (
  statistics: ApplicationInstanceStatistics,
  instances: ApplicationInstanceInfoItem[]
): string => {
  const lines: string[] = []

  lines.push(`Instance Statistics for Application: ${instances[0]?.labels?.app || 'Unknown'}`)
  lines.push(`Total: ${statistics.instanceTotal} instances, ${statistics.versionTotal} versions`)
  lines.push(`Resource Usage: CPU ${statistics.cpuTotal}, Memory ${statistics.memoryTotal}`)
  lines.push('')
  lines.push('Instance List:')
  
  instances.forEach((instance, index) => {
    lines.push(`[${index + 1}] Instance: ${instance.name} (${instance.ip})`)
    lines.push(`    Status: ${instance.deployState}, Registration: ${instance.registerState}`)
    lines.push(`    Deploy Cluster: ${instance.deployCluster || 'N/A'}`)
    lines.push(`    Labels: ${JSON.stringify(instance.labels)}`)
    lines.push('')
  })

  return lines.join('\n')
}

export const buildApplicationInstanceInfo = (
  statistics: ApplicationInstanceStatistics,
  instances: ApplicationInstanceInfoItem[]
): void => {
  setRowData('instance', { statistics, instances })
  setInfo('instance', generateApplicationInstanceInfo(statistics, instances))
}

export const generateApplicationServiceInfo = (serviceList: { serviceName: string }[]): string => {
  const lines: string[] = []
  
  lines.push(`Associated Services (${serviceList.length} total):`)
  lines.push('')
  
  serviceList.forEach((service, index) => {
    lines.push(`[${index + 1}] Service: ${service.serviceName}`)
  })

  return lines.join('\n')
}

export const buildApplicationServiceInfo = (serviceList: { serviceName: string }[]): void => {
  setRowData('service', serviceList)
  setInfo('service', generateApplicationServiceInfo(serviceList))
}

export const generateApplicationMonitorInfo = (appName: string): string => {
  return `Monitoring Dashboard for Application: ${appName}\nMetrics data is loaded from Grafana`
}

export const buildApplicationMonitorInfo = (appName: string): void => {
  setRowData('monitor', { appName })
  setInfo('monitor', generateApplicationMonitorInfo(appName))
}

export const generateApplicationTracingInfo = (appName: string): string => {
  return `Distributed Tracing for Application: ${appName}\nTrace data is loaded from Grafana`
}

export const buildApplicationTracingInfo = (appName: string): void => {
  setRowData('tracing', { appName })
  setInfo('tracing', generateApplicationTracingInfo(appName))
}

export const generateApplicationConfigInfo = (config: Record<string, unknown>): string => {
  const lines: string[] = []

  lines.push(`Configuration for Application: ${config.appName || 'Unknown'}`)

  if (config.logFlag !== undefined) {
    lines.push(`Operator Log: ${config.logFlag ? 'Enabled' : 'Disabled'}`)
  }

  if (config.rules && Array.isArray(config.rules)) {
    lines.push(`Flow Weight Rules: ${config.rules.length} configured`)
    config.rules.forEach((rule: any, index: number) => {
      lines.push(`  Rule ${index + 1}: weight=${rule.weight}`)
    })
  }

  return lines.join('\n')
}

export const buildApplicationConfigInfo = (config: Record<string, unknown>): void => {
  setRowData('config', config)
  setInfo('config', generateApplicationConfigInfo(config))
}
