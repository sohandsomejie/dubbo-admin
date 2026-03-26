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

package model

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/apache/dubbo-admin/pkg/common/constants"
	coremodel "github.com/apache/dubbo-admin/pkg/core/resource/model"
)

type ServiceSearchReq struct {
	coremodel.PageReq

	ServiceName string `form:"serviceName" json:"serviceName"`
	Keywords    string `form:"keywords" json:"keywords"`
	Mesh        string `form:"mesh" json:"mesh"`
}

func NewServiceSearchReq() *ServiceSearchReq {
	return &ServiceSearchReq{
		PageReq: coremodel.PageReq{
			PageOffset: 0,
			PageSize:   15,
		},
	}
}

type ServiceSearchResp struct {
	ServiceName     string `json:"serviceName"`
	ServiceKey      string `json:"serviceKey"`
	Version         string `json:"version"`
	Group           string `json:"group"`
	ProviderAppName string `json:"providerAppName,omitempty"`
	ConsumerAppName string `json:"consumerAppName,omitempty"`
}

type ServiceDetailResp struct {
	Language string   `json:"language"`
	Methods  []string `json:"methods"`
}

type ByServiceName []*ServiceSearchResp

func (a ByServiceName) Len() int { return len(a) }

func (a ByServiceName) Less(i, j int) bool {
	return a[i].ServiceName < a[j].ServiceName
}

func (a ByServiceName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ServiceTabDistributionReq struct {
	ServiceName     string `json:"serviceName"  form:"serviceName"`
	ServiceKeyValue string `json:"serviceKey"  form:"serviceKey"`
	Version         string `json:"version"  form:"version"`
	Group           string `json:"group"  form:"group"`
	Side            string `json:"side" form:"side"  binding:"required"`
	Mesh            string `json:"mesh" form:"mesh" binding:"required"`
	ProviderAppName string `json:"providerAppName"  form:"providerAppName"`
	Keywords        string `json:"keywords"  form:"keywords"`
	coremodel.PageReq
}

func (s *ServiceTabDistributionReq) ServiceKey() string {
	if s.ServiceKeyValue != "" {
		return s.ServiceKeyValue
	}
	return BuildServiceKey(s.ServiceName, s.Version, s.Group)
}

type ServiceTabDistributionResp struct {
	AppName      string            `json:"appName"`
	InstanceName string            `json:"instanceName"`
	Endpoint     string            `json:"endpoint"`
	TimeOut      string            `json:"timeOut"`
	Retries      string            `json:"retries"`
	Params       map[string]string `json:"params"`
}

type ByServiceInstanceName []*ServiceTabDistributionResp

func (a ByServiceInstanceName) Len() int { return len(a) }

func (a ByServiceInstanceName) Less(i, j int) bool {
	return a[i].InstanceName < a[j].InstanceName
}

func (a ByServiceInstanceName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ServiceTabDistribution struct {
	AppName      string
	InstanceName string
	Endpoint     string
	TimeOut      string
	Retries      string
}

type BaseServiceReq struct {
	ServiceName     string `json:"serviceName" form:"serviceName"`
	ServiceKeyValue string `json:"serviceKey" form:"serviceKey"`
	Group           string `json:"group" form:"group"`
	Version         string `json:"version" form:"version"`
	Mesh            string `json:"mesh" form:"mesh"`
}

func (s *BaseServiceReq) Query(c *gin.Context) error {
	s.ServiceKeyValue = c.Query("serviceKey")
	s.ServiceName = c.Query("serviceName")
	s.Group = c.Query("group")
	s.Version = c.Query("version")
	s.Mesh = c.Query("mesh")
	return s.Normalize()
}

func (s *BaseServiceReq) Normalize() error {
	if s.ServiceKeyValue != "" {
		serviceName, version, group, err := ParseServiceKey(s.ServiceKeyValue)
		if err != nil {
			return err
		}
		s.ServiceName = serviceName
		s.Version = version
		s.Group = group
		return nil
	}
	if s.ServiceName == "" {
		return fmt.Errorf("service name is empty")
	}
	s.ServiceKeyValue = BuildServiceKey(s.ServiceName, s.Version, s.Group)
	return nil
}

func (s *BaseServiceReq) ServiceKey() string {
	if s.ServiceKeyValue != "" {
		return s.ServiceKeyValue
	}
	return BuildServiceKey(s.ServiceName, s.Version, s.Group)
}

func BuildServiceKey(serviceName, version, group string) string {
	return serviceName + constants.ColonSeparator + version + constants.ColonSeparator + group
}

func ParseServiceKey(serviceKey string) (serviceName, version, group string, err error) {
	parts := strings.Split(serviceKey, constants.ColonSeparator)
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("invalid service key: %s", serviceKey)
	}
	group = parts[len(parts)-1]
	version = parts[len(parts)-2]
	serviceName = strings.Join(parts[:len(parts)-2], constants.ColonSeparator)
	if serviceName == "" {
		return "", "", "", fmt.Errorf("invalid service key: %s", serviceKey)
	}
	return serviceName, version, group, nil
}

type ServiceInterfacesReq struct {
	ServiceName string `form:"serviceName" json:"serviceName" binding:"required"`
	Mesh        string `form:"mesh" json:"mesh" binding:"required"`
}

type ServiceInterface struct {
	InterfaceName string `json:"interfaceName"`
	MethodCount   int    `json:"methodCount"`
}

type ServiceInterfacesResp struct {
	Interfaces []*ServiceInterface `json:"interfaces"`
}
