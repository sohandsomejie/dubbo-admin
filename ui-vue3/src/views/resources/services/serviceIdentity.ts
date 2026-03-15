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

export interface ServiceIdentity {
  serviceKey: string
  serviceName: string
  version: string
  group: string
}

export function buildServiceKey(serviceName: string, version = '', group = ''): string {
  return `${serviceName}:${version}:${group}`
}

export function parseServiceKey(pathId: unknown): ServiceIdentity {
  const serviceKey = Array.isArray(pathId) ? String(pathId[0] || '') : String(pathId || '')
  const segments = serviceKey.split(':')

  if (segments.length < 3) {
    return {
      serviceKey,
      serviceName: serviceKey,
      version: '',
      group: ''
    }
  }

  const group = segments.pop() || ''
  const version = segments.pop() || ''
  const serviceName = segments.join(':')

  return {
    serviceKey,
    serviceName,
    version,
    group
  }
}
