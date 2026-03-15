<!--
  ~ Licensed to the Apache Software Foundation (ASF) under one or more
  ~ contributor license agreements.  See the NOTICE file distributed with
  ~ this work for additional information regarding copyright ownership.
  ~ The ASF licenses this file to You under the Apache License, Version 2.0
  ~ (the "License"); you may not use this file except in compliance with
  ~ the License.  You may obtain a copy of the License at
  ~
  ~     http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing, software
  ~ distributed under the License is distributed on an "AS IS" BASIS,
  ~ WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  ~ See the License for the specific language governing permissions and
  ~ limitations under the License.
-->
<template>
  <div class="__container_services_tabs_detail">
    <a-card-grid>
      <a-card title="服务详情">
        <a-descriptions :column="1" layout="vertical">
          <a-descriptions-item label="服务名称">
            <p class="description-item-content">{{ serviceDetail.serviceName }}</p>
          </a-descriptions-item>
          <a-descriptions-item label="版本">
            <p class="description-item-content">{{ serviceDetail.version }}</p>
          </a-descriptions-item>
          <a-descriptions-item label="分组">
            <p class="description-item-content">{{ serviceDetail.group }}</p>
          </a-descriptions-item>
          <a-descriptions-item label="语言">
            <p class="description-item-content">{{ serviceDetail.language }}</p>
          </a-descriptions-item>
        </a-descriptions>
      </a-card>
      <a-card title="方法列表" style="margin-top: 10px">
        <a-list :data-source="serviceDetail.methods || []" size="small">
          <template #renderItem="{ item }">
            <a-list-item>{{ item }}</a-list-item>
          </template>
        </a-list>
      </a-card>
    </a-card-grid>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getServiceDetail } from '@/api/service/service'
import { parseServiceKey } from '../serviceIdentity'

const route = useRoute()

const serviceDetail = ref<{
  serviceName?: string
  version?: string
  group?: string
  language?: string
  methods?: string[]
}>({})

const onSearch = async () => {
  const serviceIdentity = parseServiceKey(route.params.pathId)
  const { data } = await getServiceDetail({
    serviceKey: serviceIdentity.serviceKey
  })
  serviceDetail.value = data.data
}

onSearch()
watch(() => route.fullPath, onSearch)
</script>

<style lang="less" scoped>
.__container_services_tabs_detail {
  .description-item-content {
    margin-left: 20px;
    width: 90%;
  }
}
</style>
