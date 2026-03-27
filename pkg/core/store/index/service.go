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

package index

import (
	"reflect"
	"strings"

	"k8s.io/client-go/tools/cache"

	"github.com/apache/dubbo-admin/pkg/common/bizerror"
	meshresource "github.com/apache/dubbo-admin/pkg/core/resource/apis/mesh/v1alpha1"
)

const ByServiceServiceName = "idx_service_service_name"
const serviceProviderAppsAnnotation = "dubbo.apache.org/provider-apps"

func init() {
	RegisterIndexers(meshresource.ServiceKind, map[string]cache.IndexFunc{
		ByServiceServiceName:     byServiceServiceName,
		ByServiceProviderAppName: byProjectedServiceProviderAppName,
	})
}

func byServiceServiceName(obj interface{}) ([]string, error) {
	svc, ok := obj.(*meshresource.ServiceResource)
	if !ok {
		return nil, bizerror.NewAssertionError(meshresource.ServiceKind, reflect.TypeOf(obj).Name())
	}
	if svc.Spec == nil {
		return []string{}, nil
	}
	return []string{svc.Spec.Name}, nil
}

func byProjectedServiceProviderAppName(obj interface{}) ([]string, error) {
	svc, ok := obj.(*meshresource.ServiceResource)
	if !ok {
		return nil, bizerror.NewAssertionError(meshresource.ServiceKind, reflect.TypeOf(obj).Name())
	}
	if svc.Annotations == nil {
		return []string{}, nil
	}
	apps := svc.Annotations[serviceProviderAppsAnnotation]
	if apps == "" {
		return []string{}, nil
	}
	return strings.Split(apps, ","), nil
}
