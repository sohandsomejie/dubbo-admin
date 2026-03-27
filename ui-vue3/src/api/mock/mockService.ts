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

import Mock from 'mockjs'
import devTool from '@/utils/DevToolUtil'

Mock.mock(devTool.mockUrl('/service/graph'), 'get', () => {
  return {
    code: 200,
    message: 'success',
    data: {
      nodes: [
        {
          id: 'provider-01',
          label: 'provider-01',
          type: 'application',
          rule: 'provider'
        },
        {
          id: 'provider-02',
          label: 'provider-02',
          type: 'application',
          rule: 'provider'
        },
        {
          id: 'serviceName:version:group',
          label: 'serviceName:version:group',
          type: 'service',
          rule: ''
        },
        {
          id: 'consumer-01',
          label: 'consumer-01',
          type: 'application',
          rule: 'consumer'
        },
        {
          id: 'consumer-02',
          label: 'consumer-02',
          type: 'application',
          rule: 'consumer'
        }
      ],
      edges: [
        {
          source: 'serviceName:version:group',
          target: 'provider-01'
        },
        {
          source: 'serviceName:version:group',
          target: 'provider-02'
        },
        {
          source: 'consumer-01',
          target: 'serviceName:version:group'
        },
        {
          source: 'consumer-02',
          target: 'serviceName:version:group'
        }
      ]
    }
  }
})

Mock.mock(devTool.mockUrl('/mock/service/search'), 'get', {
  code: 200,
  message: 'success',
  data: {
    pageInfo: {
      total: 8,
      pageSize: 15,
      pageOffset: 0
    },
    list: [
      {
        serviceName: 'org.apache.dubbo.samples.UserService',
        serviceKey: 'org.apache.dubbo.samples.UserService:1.0.0:group1',
        version: '1.0.0',
        group: 'group1'
      },
      {
        serviceName: 'org.apache.dubbo.samples.UserService',
        serviceKey: 'org.apache.dubbo.samples.UserService:1.0.0:',
        version: '1.0.0',
        group: ''
      },
      {
        serviceName: 'org.apache.dubbo.samples.OrderService',
        serviceKey: 'org.apache.dubbo.samples.OrderService:1.0.0:group1',
        version: '1.0.0',
        group: 'group1'
      },
      {
        serviceName: 'org.apache.dubbo.samples.DetailService',
        serviceKey: 'org.apache.dubbo.samples.DetailService:1.0.0:group1',
        version: '1.0.0',
        group: 'group1'
      },
      {
        serviceName: 'org.apache.dubbo.samples.PayService',
        serviceKey: 'org.apache.dubbo.samples.PayService:1.0.0:group1',
        version: '1.0.0',
        group: 'group1'
      },
      {
        serviceName: 'org.apache.dubbo.samples.CommentService',
        serviceKey: 'org.apache.dubbo.samples.CommentService:1.0.0:group1',
        version: '1.0.0',
        group: 'group1'
      },
      {
        serviceName: 'org.apache.dubbo.samples.RepayService',
        serviceKey: 'org.apache.dubbo.samples.RepayService:1.0.0:group1',
        version: '1.0.0',
        group: 'group1'
      },
      {
        serviceName: 'org.apche.dubbo.samples.TransportService',
        serviceKey: 'org.apche.dubbo.samples.TransportService:1.0.0:group1',
        version: '1.0.0',
        group: 'group1'
      }
    ]
  }
})
