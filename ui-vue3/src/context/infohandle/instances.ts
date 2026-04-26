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

export interface InstanceSearchItem {
  ip: string
  name: string
  deployState: string
  deployCluster: string
  registerState: string
  registerClusters: string[]
  cpu: string
  memory: string
  startTime_k8s: string
  registerTime: string
  labels: Record<string, string>
  appName?: string
  lifecycleState?: string
}

export interface InstanceDetail {
  name: string
  deployState: string
  registerStates: string
  ip: string
  rpcPort: string
  appName: string
  workloadName: string
  labels: Record<string, string>
  createTime: string
  readyTime: string
  registerTime: string
  registerClusters: string[]
  deployCluster: string
  node: string
  image: string
  probes: {
    startupProbe: { type: string; open: boolean }
    readinessProbe: { type: string; open: boolean }
    livenessProbe: { type: string; open: boolean }
  }
}

export const generateInstancesListInfo = (data: InstanceSearchItem[]): string => {
  const lines: string[] = []
  
  lines.push(`Instances List (${data.length} total):`)
  lines.push('')
  
  data.forEach((item, index) => {
    lines.push(`[${index + 1}] Instance: ${item.name} (${item.ip})`)
    lines.push(`    Application: ${item.appName || 'N/A'}`)
    lines.push(`    Status: ${item.deployState}, Registration: ${item.registerState}`)
    lines.push(`    Deploy Cluster: ${item.deployCluster || 'N/A'}`)
    lines.push(`    Labels: ${JSON.stringify(item.labels)}`)
    lines.push('')
  })

  return lines.join('\n')
}

export const buildInstancesListInfo = (data: InstanceSearchItem[]): void => {
  setRowData('list', data)
  setInfo('list', generateInstancesListInfo(data))
}

export const generateInstanceDetailInfo = (data: InstanceDetail): string => {
  const lines: string[] = []

  lines.push(`Instance: ${data.name || data.ip}`)
  lines.push(`IP Address: ${data.ip}`)
  lines.push(`Application: ${data.appName}`)
  lines.push(`RPC Port: ${data.rpcPort || 'N/A'}`)
  lines.push('')
  lines.push(`Deployment Status: ${data.deployState}`)
  lines.push(`Registration Status: ${data.registerStates}`)
  lines.push(`Deploy Cluster: ${data.deployCluster || 'N/A'}`)
  lines.push(`Node: ${data.node || 'N/A'}`)
  
  if (data.registerClusters && data.registerClusters.length > 0) {
    lines.push(`Registry Clusters: ${data.registerClusters.join(', ')}`)
  }
  
  lines.push(`Workload: ${data.workloadName || 'N/A'}`)
  lines.push(`Container Image: ${data.image || 'N/A'}`)
  lines.push('')
  lines.push(`Timeline:`)
  lines.push(`    Created: ${data.createTime || 'N/A'}`)
  lines.push(`    Ready: ${data.readyTime || 'N/A'}`)
  lines.push(`    Registered: ${data.registerTime || 'N/A'}`)
  
  if (data.probes) {
    lines.push('')
    lines.push(`Health Probes:`)
    lines.push(`    Startup Probe: ${data.probes.startupProbe?.open ? 'Enabled' : 'Disabled'} (${data.probes.startupProbe?.type || 'N/A'})`)
    lines.push(`    Readiness Probe: ${data.probes.readinessProbe?.open ? 'Enabled' : 'Disabled'} (${data.probes.readinessProbe?.type || 'N/A'})`)
    lines.push(`    Liveness Probe: ${data.probes.livenessProbe?.open ? 'Enabled' : 'Disabled'} (${data.probes.livenessProbe?.type || 'N/A'})`)
  }

  if (data.labels && Object.keys(data.labels).length > 0) {
    lines.push('')
    lines.push(`Labels: ${JSON.stringify(data.labels)}`)
  }

  return lines.join('\n')
}

export const buildInstanceDetailInfo = (data: InstanceDetail): void => {
  setRowData('detail', data)
  setInfo('detail', generateInstanceDetailInfo(data))
}

export const generateInstanceMonitorInfo = (instanceName: string, appName?: string): string => {
  const lines: string[] = []
  lines.push(`Monitoring Dashboard for Instance: ${instanceName}`)
  if (appName) {
    lines.push(`Application: ${appName}`)
  }
  lines.push('')
  lines.push(`Metrics data is loaded from Grafana`)
  return lines.join('\n')
}

export const buildInstanceMonitorInfo = (instanceName: string, appName?: string): void => {
  setRowData('monitor', { instanceName, appName })
  setInfo('monitor', generateInstanceMonitorInfo(instanceName, appName))
}

export const generateInstanceLinkTrackingInfo = (instanceName: string, appName?: string): string => {
  const lines: string[] = []
  lines.push(`Link Tracking for Instance: ${instanceName}`)
  if (appName) {
    lines.push(`Application: ${appName}`)
  }
  lines.push('')
  lines.push(`Trace data is loaded from Grafana`)
  return lines.join('\n')
}

export const buildInstanceLinkTrackingInfo = (instanceName: string, appName?: string): void => {
  setRowData('linkTracking', { instanceName, appName })
  setInfo('linkTracking', generateInstanceLinkTrackingInfo(instanceName, appName))
}

export const generateInstanceConfigurationInfo = (config: Record<string, unknown>): string => {
  const lines: string[] = []

  lines.push(`Configuration for Instance: ${config.name || config.ip || 'Unknown'}`)
  if (config.appName) {
    lines.push(`Application: ${config.appName}`)
  }
  lines.push('')
  
  if (config.logSwitch !== undefined) {
    lines.push(`Operator Log: ${config.logSwitch ? 'Enabled' : 'Disabled'}`)
  }

  if (config.trafficSwitch !== undefined) {
    lines.push(`Traffic Disable: ${config.trafficSwitch ? 'Disabled' : 'Enabled'}`)
  }

  return lines.join('\n')
}

export const buildInstanceConfigurationInfo = (config: Record<string, unknown>): void => {
  setRowData('configuration', config)
  setInfo('configuration', generateInstanceConfigurationInfo(config))
}
