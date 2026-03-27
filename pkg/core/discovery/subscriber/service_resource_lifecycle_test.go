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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/cache"

	meshproto "github.com/apache/dubbo-admin/api/mesh/v1alpha1"
	"github.com/apache/dubbo-admin/pkg/core/events"
	meshresource "github.com/apache/dubbo-admin/pkg/core/resource/apis/mesh/v1alpha1"
	coremodel "github.com/apache/dubbo-admin/pkg/core/resource/model"
	corestore "github.com/apache/dubbo-admin/pkg/core/store"
	memorystore "github.com/apache/dubbo-admin/pkg/store/memory"
)

// noopEmitter discards all emitted events; it is safe for use in unit tests.
type noopEmitter struct{}

func (n *noopEmitter) Send(_ events.Event) {}

// newTestStores creates initialised, in-memory app and service stores.
func newTestStores(t *testing.T) (appStore, serviceStore corestore.ResourceStore) {
	t.Helper()
	appS := memorystore.NewMemoryResourceStore(meshresource.ApplicationKind)
	require.NoError(t, appS.Init(nil))
	svcS := memorystore.NewMemoryResourceStore(meshresource.ServiceKind)
	require.NoError(t, svcS.Init(nil))
	return appS, svcS
}

func newProviderMetadata(serviceName, version, group, providerApp, mesh string, methodNames ...string) *meshresource.ServiceProviderMetadataResource {
	res := meshresource.NewServiceProviderMetadataResourceWithAttributes(
		meshresource.BuildServiceKey(serviceName, version, group, providerApp), mesh)
	methods := make([]*meshproto.Method, 0, len(methodNames))
	for _, name := range methodNames {
		methods = append(methods, &meshproto.Method{Name: name})
	}
	res.Spec = &meshproto.ServiceProviderMetadata{
		ServiceName:     serviceName,
		Version:         version,
		Group:           group,
		ProviderAppName: providerApp,
		Methods:         methods,
	}
	return res
}

func newDeletedProviderEvent(res *meshresource.ServiceProviderMetadataResource) events.Event {
	return events.NewResourceChangedEvent(cache.Deleted, res, nil)
}

// TestProviderUpsert_CreatesServiceWithMethods verifies that upserting a provider
// metadata creates a ServiceResource with the correct methods from the spec.
func TestProviderUpsert_CreatesServiceWithMethods(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	const (
		mesh        = ""
		serviceName = "com.example.DemoService"
		version     = "1.0"
		group       = "grp"
		providerApp = "demo-provider"
	)

	provMeta := newProviderMetadata(serviceName, version, group, providerApp, mesh, "sayHello", "getUserById")
	err := sub.ProcessEvent(events.NewResourceChangedEvent(cache.Added, nil, provMeta))
	require.NoError(t, err)

	svcKey := serviceName + ":1.0:grp"
	raw, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, svcKey))
	require.NoError(t, err)
	require.True(t, exists, "ServiceResource should be created")

	svcRes := raw.(*meshresource.ServiceResource)
	assert.ElementsMatch(t, []string{"sayHello", "getUserById"}, svcRes.Spec.Methods)
}

// TestProviderUpsert_MergesMethodsFromMultipleProviders verifies that upserting
// two providers for the same service key produces a union of their methods.
func TestProviderUpsert_MergesMethodsFromMultipleProviders(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	const (
		mesh        = ""
		serviceName = "com.example.DemoService"
		version     = "1.0"
		group       = "grp"
	)

	provMeta1 := newProviderMetadata(serviceName, version, group, "provider-a", mesh, "sayHello", "getUserById")
	require.NoError(t, sub.ProcessEvent(events.NewResourceChangedEvent(cache.Added, nil, provMeta1)))
	require.NoError(t, providerStore.Add(provMeta1))

	provMeta2 := newProviderMetadata(serviceName, version, group, "provider-b", mesh, "getUserById", "listUsers")
	require.NoError(t, sub.ProcessEvent(events.NewResourceChangedEvent(cache.Added, nil, provMeta2)))
	require.NoError(t, providerStore.Add(provMeta2))

	svcKey := serviceName + ":1.0:grp"
	raw, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, svcKey))
	require.NoError(t, err)
	require.True(t, exists)

	svcRes := raw.(*meshresource.ServiceResource)
	assert.ElementsMatch(t, []string{"sayHello", "getUserById", "listUsers"}, svcRes.Spec.Methods)
}

// TestProviderDelete_LastProvider_DeletesServiceResource verifies that deleting the
// last provider removes the ServiceResource.
func TestProviderDelete_LastProvider_DeletesServiceResource(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	const (
		mesh        = ""
		serviceName = "com.example.DemoService"
		version     = "1.0"
		group       = "grp"
		providerApp = "demo-provider"
	)

	// Pre-populate: add the service resource
	svcKey := serviceName + ":1.0:grp"
	svcRes := meshresource.NewServiceResourceWithAttributes(svcKey, mesh)
	svcRes.Spec.Name = serviceName
	svcRes.Spec.Version = version
	svcRes.Spec.Group = group
	svcRes.Spec.Methods = []string{"sayHello"}
	require.NoError(t, svcStore.Add(svcRes))

	provMeta := newProviderMetadata(serviceName, version, group, providerApp, mesh)
	require.NoError(t, providerStore.Add(provMeta))
	err := sub.ProcessEvent(newDeletedProviderEvent(provMeta))
	require.NoError(t, err)

	_, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, svcKey))
	require.NoError(t, err)
	assert.False(t, exists, "ServiceResource should be deleted when provider is removed")
}

func TestProviderDelete_RebuildsMethodsFromRemainingProviders(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	const (
		mesh        = ""
		serviceName = "com.example.DemoService"
		version     = "1.0"
		group       = "grp"
	)

	provMeta1 := newProviderMetadata(serviceName, version, group, "provider-a", mesh, "sayHello")
	provMeta2 := newProviderMetadata(serviceName, version, group, "provider-b", mesh, "listUsers")
	require.NoError(t, providerStore.Add(provMeta1))
	require.NoError(t, providerStore.Add(provMeta2))

	svcKey := serviceName + ":1.0:grp"
	svcRes := meshresource.NewServiceResourceWithAttributes(svcKey, mesh)
	svcRes.Spec.Name = serviceName
	svcRes.Spec.Version = version
	svcRes.Spec.Group = group
	svcRes.Spec.Methods = []string{"sayHello", "listUsers"}
	require.NoError(t, svcStore.Add(svcRes))

	err := sub.ProcessEvent(newDeletedProviderEvent(provMeta1))
	require.NoError(t, err)

	raw, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, svcKey))
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, []string{"listUsers"}, raw.(*meshresource.ServiceResource).Spec.Methods)
	assert.Equal(t, "provider-b", raw.(*meshresource.ServiceResource).Annotations[serviceProviderAppsAnnotation])
}

// TestProviderUpdate_ServiceKeyChanged_CleansOldRelationship verifies that when
// a provider metadata Updated event changes the service key, the old ServiceResource
// is deleted and a new one is created.
func TestProviderUpdate_ServiceKeyChanged_CleansOldRelationship(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	const (
		mesh        = ""
		providerApp = "demo-provider"
	)

	// Pre-populate old service resource.
	oldSvcKey := "com.example.OldService:1.0:grp"
	oldSvcRes := meshresource.NewServiceResourceWithAttributes(oldSvcKey, mesh)
	oldSvcRes.Spec.Name = "com.example.OldService"
	oldSvcRes.Spec.Methods = []string{"oldMethod"}
	require.NoError(t, svcStore.Add(oldSvcRes))

	oldMeta := newProviderMetadata("com.example.OldService", "1.0", "grp", providerApp, mesh)
	newMeta := newProviderMetadata("com.example.NewService", "1.0", "grp", providerApp, mesh, "newMethod")
	require.NoError(t, providerStore.Add(oldMeta))
	updateEvent := events.NewResourceChangedEvent(cache.Updated, oldMeta, newMeta)

	err := sub.ProcessEvent(updateEvent)
	require.NoError(t, err)

	// Old service resource should be gone.
	_, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, oldSvcKey))
	require.NoError(t, err)
	assert.False(t, exists, "Old ServiceResource should be deleted after provider moved to new service")

	// New service resource should exist with the method.
	newSvcKey := "com.example.NewService:1.0:grp"
	raw, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, newSvcKey))
	require.NoError(t, err)
	require.True(t, exists, "New ServiceResource should be created")
	assert.Equal(t, []string{"newMethod"}, raw.(*meshresource.ServiceResource).Spec.Methods)
}

// TestProviderUpdate_SameKey_NoStaleCleanup verifies that a no-op Updated event
// (same service key) does not delete the Service.
func TestProviderUpdate_SameKey_NoStaleCleanup(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	const (
		mesh        = ""
		serviceName = "com.example.DemoService"
		version     = "1.0"
		group       = "grp"
		providerApp = "demo-provider"
	)

	svcKey := serviceName + ":1.0:grp"
	svcRes := meshresource.NewServiceResourceWithAttributes(svcKey, mesh)
	svcRes.Spec.Name = serviceName
	svcRes.Spec.Methods = []string{"sayHello"}
	require.NoError(t, svcStore.Add(svcRes))

	meta := newProviderMetadata(serviceName, version, group, providerApp, mesh, "sayHello")
	updateEvent := events.NewResourceChangedEvent(cache.Updated, meta, meta)

	err := sub.ProcessEvent(updateEvent)
	require.NoError(t, err)

	raw, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, svcKey))
	require.NoError(t, err)
	require.True(t, exists, "ServiceResource should still exist after no-op update")
	assert.Contains(t, raw.(*meshresource.ServiceResource).Spec.Methods, "sayHello")
}

func TestProviderUpdate_SameKey_RebuildsMethodsAndProviderApps(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	const (
		mesh        = ""
		serviceName = "com.example.DemoService"
		version     = "1.0"
		group       = "grp"
	)

	oldMeta := newProviderMetadata(serviceName, version, group, "provider-a", mesh, "sayHello", "oldMethod")
	newMeta := newProviderMetadata(serviceName, version, group, "provider-b", mesh, "sayHello", "listUsers")
	require.NoError(t, providerStore.Add(oldMeta))

	err := sub.ProcessEvent(events.NewResourceChangedEvent(cache.Updated, oldMeta, newMeta))
	require.NoError(t, err)

	svcKey := serviceName + ":1.0:grp"
	raw, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, svcKey))
	require.NoError(t, err)
	require.True(t, exists)

	svcRes := raw.(*meshresource.ServiceResource)
	assert.Equal(t, []string{"listUsers", "sayHello"}, svcRes.Spec.Methods)
	assert.Equal(t, "provider-b", svcRes.Annotations[serviceProviderAppsAnnotation])
}

func TestProviderUpsert_StoresProviderAppsAnnotation(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	const (
		mesh        = ""
		serviceName = "com.example.DemoService"
		version     = "1.0"
		group       = "grp"
	)

	provMeta1 := newProviderMetadata(serviceName, version, group, "provider-a", mesh, "sayHello")
	provMeta2 := newProviderMetadata(serviceName, version, group, "provider-b", mesh, "listUsers")
	require.NoError(t, sub.ProcessEvent(events.NewResourceChangedEvent(cache.Added, nil, provMeta1)))
	require.NoError(t, providerStore.Add(provMeta1))
	require.NoError(t, sub.ProcessEvent(events.NewResourceChangedEvent(cache.Added, nil, provMeta2)))

	svcKey := serviceName + ":1.0:grp"
	raw, exists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, svcKey))
	require.NoError(t, err)
	require.True(t, exists)

	apps := strings.Split(raw.(*meshresource.ServiceResource).Annotations[serviceProviderAppsAnnotation], ",")
	assert.ElementsMatch(t, []string{"provider-a", "provider-b"}, apps)
}

func TestProviderUpdate_NewSpecNil_ReturnsErrorWithoutPanic(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	providerStore := memorystore.NewMemoryResourceStore(meshresource.ServiceProviderMetadataKind)
	require.NoError(t, providerStore.Init(nil))
	sub := NewServiceProviderMetadataEventSubscriber(appStore, svcStore, providerStore, emitter)

	oldMeta := newProviderMetadata("com.example.DemoService", "1.0", "grp", "demo-provider", "")
	newMeta := meshresource.NewServiceProviderMetadataResourceWithAttributes("com.example.DemoService", "")
	newMeta.Spec = nil
	updateEvent := events.NewResourceChangedEvent(cache.Updated, oldMeta, newMeta)

	require.NotPanics(t, func() {
		err := sub.ProcessEvent(updateEvent)
		require.EqualError(t, err, "[UnknownError], provider metadata resource spec is nil")
	})
}

// TestConsumerUpsert_CreatesApplication verifies that the consumer subscriber only
// creates an Application resource and does NOT touch the Service store.
func TestConsumerUpsert_CreatesApplication(t *testing.T) {
	appStore, svcStore := newTestStores(t)
	emitter := &noopEmitter{}
	sub := NewServiceConsumerMetadataEventSubscriber(appStore, emitter)

	const (
		mesh        = ""
		serviceName = "com.example.DemoService"
		version     = "1.0"
		group       = "grp"
		consumerApp = "demo-consumer"
	)

	consMeta := meshresource.NewServiceConsumerMetadataResourceWithAttributes(serviceName, mesh)
	consMeta.Spec = &meshproto.ServiceConsumerMetadata{
		ServiceName:     serviceName,
		Version:         version,
		Group:           group,
		ConsumerAppName: consumerApp,
	}

	err := sub.ProcessEvent(events.NewResourceChangedEvent(cache.Added, nil, consMeta))
	require.NoError(t, err)

	// Application should exist.
	_, exists, err := appStore.GetByKey(coremodel.BuildResourceKey(mesh, consumerApp))
	require.NoError(t, err)
	assert.True(t, exists, "Application should be created by consumer subscriber")

	// Service store should be untouched.
	svcKey := serviceName + ":1.0:grp"
	_, svcExists, err := svcStore.GetByKey(coremodel.BuildResourceKey(mesh, svcKey))
	require.NoError(t, err)
	assert.False(t, svcExists, "Consumer subscriber must NOT write to the Service store")
}
