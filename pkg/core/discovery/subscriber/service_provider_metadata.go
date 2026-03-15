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

package subscriber

import (
	"reflect"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"k8s.io/client-go/tools/cache"

	meshproto "github.com/apache/dubbo-admin/api/mesh/v1alpha1"
	"github.com/apache/dubbo-admin/pkg/common/bizerror"
	"github.com/apache/dubbo-admin/pkg/common/constants"
	"github.com/apache/dubbo-admin/pkg/core/events"
	"github.com/apache/dubbo-admin/pkg/core/logger"
	meshresource "github.com/apache/dubbo-admin/pkg/core/resource/apis/mesh/v1alpha1"
	coremodel "github.com/apache/dubbo-admin/pkg/core/resource/model"
	"github.com/apache/dubbo-admin/pkg/core/store"
	"github.com/apache/dubbo-admin/pkg/core/store/index"
)

type ServiceProviderMetadataEventSubscriber struct {
	appStore             store.ResourceStore
	serviceStore         store.ResourceStore
	serviceProviderStore store.ResourceStore
	emitter              events.Emitter
}

func NewServiceProviderMetadataEventSubscriber(
	appStore store.ResourceStore,
	serviceStore store.ResourceStore,
	serviceProviderStore store.ResourceStore,
	emitter events.Emitter) *ServiceProviderMetadataEventSubscriber {
	return &ServiceProviderMetadataEventSubscriber{
		appStore:             appStore,
		serviceStore:         serviceStore,
		serviceProviderStore: serviceProviderStore,
		emitter:              emitter,
	}
}

func (s *ServiceProviderMetadataEventSubscriber) ResourceKind() coremodel.ResourceKind {
	return meshresource.ServiceProviderMetadataKind
}

func (s *ServiceProviderMetadataEventSubscriber) Name() string {
	return "Discovery-" + s.ResourceKind().ToString()
}

func (s *ServiceProviderMetadataEventSubscriber) ProcessEvent(event events.Event) error {
	newObj, ok := event.NewObj().(*meshresource.ServiceProviderMetadataResource)
	if !ok && event.NewObj() != nil {
		return bizerror.NewAssertionError(reflect.TypeOf(newObj), event.NewObj())
	}
	var processErr error
	switch event.Type() {
	case cache.Added, cache.Replaced, cache.Sync:
		if newObj == nil {
			errStr := "process provider metadata resource upsert event, but new obj is nil, skipped processing"
			logger.Errorf(errStr)
			return bizerror.New(bizerror.EventError, errStr)
		}
		processErr = s.processUpsert(newObj)
	case cache.Updated:
		if newObj == nil {
			errStr := "process provider metadata resource update event, but new obj is nil, skipped processing"
			logger.Errorf(errStr)
			return bizerror.New(bizerror.EventError, errStr)
		}
		oldObj, _ := event.OldObj().(*meshresource.ServiceProviderMetadataResource)
		processErr = s.processUpdate(oldObj, newObj)
	case cache.Deleted:
		oldObj, ok := event.OldObj().(*meshresource.ServiceProviderMetadataResource)
		if !ok && event.OldObj() != nil {
			return bizerror.NewAssertionError(reflect.TypeOf(oldObj), event.OldObj())
		}
		if oldObj != nil {
			processErr = s.processDelete(oldObj)
		} else {
			logger.Warnf("provider metadata deleted event has nil OldObj, skipping")
		}
	}
	if processErr != nil {
		logger.Errorf("process provider metadata resource event failed, cause: %s, event: %s", processErr.Error(), event.String())
		return processErr
	}
	logger.Infof("process provider metadata resource event successfully, event: %s", event.String())
	return nil
}

func (s *ServiceProviderMetadataEventSubscriber) processUpsert(r *meshresource.ServiceProviderMetadataResource) error {
	if r.Spec == nil {
		return bizerror.New(bizerror.UnknownError, "provider metadata resource spec is nil")
	}
	if strutil.IsBlank(r.Spec.ProviderAppName) {
		logger.Warnf("skip processing service provider metadata event because spec.providerAppName is blank, res:%s", r.String())
		return nil
	}

	// upsert Application
	_, exists, err := s.appStore.GetByKey(coremodel.BuildResourceKey(r.Mesh, r.Spec.ProviderAppName))
	if err != nil {
		logger.Errorf("get application resource failed, appName: %s, mesh: %s, cause: %s",
			r.Spec.ProviderAppName, r.Mesh, err.Error())
		return err
	}
	if !exists {
		appRes := meshresource.NewApplicationResourceWithAttributes(r.Spec.ProviderAppName, r.Mesh)
		appRes.Spec.Name = r.Spec.ProviderAppName
		if err := s.appStore.Add(appRes); err != nil {
			logger.Errorf("add application resource failed, appName: %s, mesh: %s, cause: %s",
				r.Spec.ProviderAppName, r.Mesh, err.Error())
			return err
		}
		s.emitter.Send(events.NewResourceChangedEvent(cache.Added, nil, appRes))
	} else {
		logger.Infof("application resource already exists, appName: %s, mesh: %s", r.Spec.ProviderAppName, r.Mesh)
	}

	// upsert Service (sync methods)
	return s.upsertService(r)
}

func (s *ServiceProviderMetadataEventSubscriber) upsertService(r *meshresource.ServiceProviderMetadataResource) error {
	if strutil.IsBlank(r.Spec.ServiceName) {
		logger.Warnf("skip service upsert because spec.serviceName is blank, res:%s", r.String())
		return nil
	}
	svcName := r.Spec.ServiceName + constants.ColonSeparator + r.Spec.Version + constants.ColonSeparator + r.Spec.Group
	svcKey := coremodel.BuildResourceKey(r.Mesh, svcName)

	// Extract method names from provider metadata
	newMethods := extractMethodNames(r.Spec.Methods)

	raw, exists, err := s.serviceStore.GetByKey(svcKey)
	if err != nil {
		logger.Errorf("get service resource failed, svcKey: %s, cause: %s", svcKey, err.Error())
		return err
	}
	if exists {
		svcRes, ok := raw.(*meshresource.ServiceResource)
		if !ok {
			return bizerror.NewAssertionError(meshresource.ServiceKind, raw)
		}
		merged := slice.Unique(append(append([]string{}, svcRes.Spec.Methods...), newMethods...))
		svcRes.Spec.Methods = merged
		if err := s.serviceStore.Update(svcRes); err != nil {
			logger.Errorf("update service resource failed, svcKey: %s, cause: %s", svcKey, err.Error())
			return err
		}
		s.emitter.Send(events.NewResourceChangedEvent(cache.Updated, nil, svcRes))
		return nil
	}

	svcRes := meshresource.NewServiceResourceWithAttributes(svcName, r.Mesh)
	svcRes.Spec.Name = r.Spec.ServiceName
	svcRes.Spec.Version = r.Spec.Version
	svcRes.Spec.Group = r.Spec.Group
	svcRes.Spec.Methods = newMethods
	if err := s.serviceStore.Add(svcRes); err != nil {
		logger.Errorf("add service resource failed, svcKey: %s, cause: %s", svcKey, err.Error())
		return err
	}
	s.emitter.Send(events.NewResourceChangedEvent(cache.Added, nil, svcRes))
	return nil
}

// extractMethodNames extracts the Name from each Method proto message.
func extractMethodNames(methods []*meshproto.Method) []string {
	names := make([]string, 0, len(methods))
	for _, m := range methods {
		if m != nil && m.Name != "" {
			names = append(names, m.Name)
		}
	}
	return names
}

func (s *ServiceProviderMetadataEventSubscriber) processUpdate(
	oldObj *meshresource.ServiceProviderMetadataResource,
	newObj *meshresource.ServiceProviderMetadataResource) error {
	if newObj.Spec == nil {
		return s.processUpsert(newObj)
	}

	// If the service key changed, delete the old Service resource first.
	if oldObj != nil && oldObj.Spec != nil && !strutil.IsBlank(oldObj.Spec.ServiceName) {
		oldSvcName := oldObj.Spec.ServiceName + constants.ColonSeparator + oldObj.Spec.Version + constants.ColonSeparator + oldObj.Spec.Group
		newSvcName := newObj.Spec.ServiceName + constants.ColonSeparator + newObj.Spec.Version + constants.ColonSeparator + newObj.Spec.Group
		if oldSvcName != newSvcName {
			logger.Infof("provider metadata service key changed (old: %s, new: %s), cleaning up stale service",
				oldSvcName, newSvcName)
			if err := s.processDelete(oldObj); err != nil {
				return err
			}
		}
	}
	return s.processUpsert(newObj)
}

func (s *ServiceProviderMetadataEventSubscriber) processDelete(r *meshresource.ServiceProviderMetadataResource) error {
	if r.Spec == nil {
		return nil
	}
	if strutil.IsBlank(r.Spec.ServiceName) {
		return nil
	}
	svcName := r.Spec.ServiceName + constants.ColonSeparator + r.Spec.Version + constants.ColonSeparator + r.Spec.Group
	svcKey := coremodel.BuildResourceKey(r.Mesh, svcName)

	raw, exists, err := s.serviceStore.GetByKey(svcKey)
	if err != nil {
		logger.Errorf("get service resource failed during delete, svcKey: %s, cause: %s", svcKey, err.Error())
		return err
	}
	if !exists {
		return nil
	}
	svcRes, ok := raw.(*meshresource.ServiceResource)
	if !ok {
		return bizerror.NewAssertionError(meshresource.ServiceKind, raw)
	}

	methods, err := s.collectMethodsForService(r.Mesh, svcName, svcKey, r.ResourceKey())
	if err != nil {
		return err
	}
	if len(methods) == 0 {
		if err := s.serviceStore.Delete(svcRes); err != nil {
			logger.Errorf("delete service resource failed, svcKey: %s, cause: %s", svcKey, err.Error())
			return err
		}
		s.emitter.Send(events.NewResourceChangedEvent(cache.Deleted, svcRes, nil))
		return nil
	}

	svcRes.Spec.Methods = methods
	if err := s.serviceStore.Update(svcRes); err != nil {
		logger.Errorf("update service resource failed during delete, svcKey: %s, cause: %s", svcKey, err.Error())
		return err
	}
	s.emitter.Send(events.NewResourceChangedEvent(cache.Updated, nil, svcRes))
	return nil
}

func (s *ServiceProviderMetadataEventSubscriber) collectMethodsForService(mesh, svcName, svcKey, excludeKey string) ([]string, error) {
	resources, err := s.serviceProviderStore.ListByIndexes(map[string]string{
		index.ByServiceProviderServiceKey: svcName,
		index.ByMeshIndex:                 mesh,
	})
	if err != nil {
		logger.Errorf("list provider metadata failed, svcKey: %s, cause: %s", svcKey, err.Error())
		return nil, err
	}

	methods := make([]string, 0)
	for _, resource := range resources {
		provider, ok := resource.(*meshresource.ServiceProviderMetadataResource)
		if !ok {
			return nil, bizerror.NewAssertionError(meshresource.ServiceProviderMetadataKind, resource)
		}
		if provider.ResourceKey() == excludeKey || provider.Spec == nil {
			continue
		}
		methods = append(methods, extractMethodNames(provider.Spec.Methods)...)
	}
	return slice.Unique(methods), nil
}
