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

package service

import (
	ctx "context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apache/dubbo-admin/pkg/common/constants"
	"github.com/apache/dubbo-admin/pkg/config/app"
	discoverycfg "github.com/apache/dubbo-admin/pkg/config/discovery"
	consolectx "github.com/apache/dubbo-admin/pkg/console/context"
	"github.com/apache/dubbo-admin/pkg/console/counter"
	"github.com/apache/dubbo-admin/pkg/console/model"
	"github.com/apache/dubbo-admin/pkg/core/manager"
	meshresource "github.com/apache/dubbo-admin/pkg/core/resource/apis/mesh/v1alpha1"
	coremodel "github.com/apache/dubbo-admin/pkg/core/resource/model"
	corestore "github.com/apache/dubbo-admin/pkg/core/store"
	memorystore "github.com/apache/dubbo-admin/pkg/store/memory"
)

type testStoreRouter struct {
	stores map[coremodel.ResourceKind]corestore.ResourceStore
}

func (r *testStoreRouter) ResourceRoute(res coremodel.Resource) (corestore.ResourceStore, error) {
	return r.ResourceKindRoute(res.ResourceKind())
}

func (r *testStoreRouter) ResourceKindRoute(kind coremodel.ResourceKind) (corestore.ResourceStore, error) {
	store, ok := r.stores[kind]
	if !ok {
		return nil, fmt.Errorf("resource kind %s not found", kind)
	}
	return store, nil
}

type testConsoleContext struct {
	rm  manager.ResourceManager
	cfg app.AdminConfig
}

func (c *testConsoleContext) ResourceManager() manager.ResourceManager {
	return c.rm
}

func (c *testConsoleContext) CounterManager() counter.CounterManager {
	return nil
}

func (c *testConsoleContext) Config() app.AdminConfig {
	return c.cfg
}

func (c *testConsoleContext) AppContext() ctx.Context {
	return ctx.Background()
}

func newTestConsoleContext(t *testing.T) (consolectx.Context, corestore.ResourceStore, corestore.ResourceStore) {
	t.Helper()

	appStore := memorystore.NewMemoryResourceStore(meshresource.ApplicationKind)
	require.NoError(t, appStore.Init(nil))

	serviceStore := memorystore.NewMemoryResourceStore(meshresource.ServiceKind)
	require.NoError(t, serviceStore.Init(nil))

	rm := manager.NewResourceManager(&testStoreRouter{
		stores: map[coremodel.ResourceKind]corestore.ResourceStore{
			meshresource.ApplicationKind: appStore,
			meshresource.ServiceKind:     serviceStore,
		},
	}, nil)

	cfg := app.DefaultAdminConfig()
	cfg.Engine.Name = "test-engine"
	cfg.Discovery = []*discoverycfg.Config{
		{ID: "test-mesh", Name: "test-registry"},
	}

	return &testConsoleContext{
		rm:  rm,
		cfg: cfg,
	}, appStore, serviceStore
}

func newServiceResource(mesh, serviceName, version, group string, methods []string) *meshresource.ServiceResource {
	serviceKey := model.BuildServiceKey(serviceName, version, group)
	res := meshresource.NewServiceResourceWithAttributes(serviceKey, mesh)
	res.Spec.Name = serviceName
	res.Spec.Version = version
	res.Spec.Group = group
	res.Spec.Methods = methods
	return res
}

func newServiceResourceForApp(mesh, appName, serviceName, version, group string, methods []string) *meshresource.ServiceResource {
	res := newServiceResource(mesh, serviceName, version, group, methods)
	res.Annotations = map[string]string{
		"dubbo.apache.org/provider-apps": appName,
	}
	return res
}

func TestSearchServices_ReturnsCanonicalServiceKeys(t *testing.T) {
	consoleCtx, _, serviceStore := newTestConsoleContext(t)
	const mesh = "test-mesh"

	require.NoError(t, serviceStore.Add(newServiceResource(mesh,
		"org.apache.dubbo.samples.UserService", "1.0.0", "gray",
		[]string{"getUserById", "listUsers"})))
	require.NoError(t, serviceStore.Add(newServiceResource(mesh,
		"org.apache.dubbo.samples.UserService", "2.0.0", "gray",
		[]string{"getUserById", "createUser"})))

	resp, err := SearchServices(consoleCtx, &model.ServiceSearchReq{
		Mesh: mesh,
		PageReq: coremodel.PageReq{
			PageOffset: 0,
			PageSize:   10,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	list, ok := resp.List.([]*model.ServiceSearchResp)
	require.True(t, ok)
	require.Len(t, list, 2)

	byKey := map[string]*model.ServiceSearchResp{}
	for _, item := range list {
		byKey[item.ServiceKey] = item
	}

	key1 := model.BuildServiceKey("org.apache.dubbo.samples.UserService", "1.0.0", "gray")
	key2 := model.BuildServiceKey("org.apache.dubbo.samples.UserService", "2.0.0", "gray")
	assert.Equal(t, "1.0.0", byKey[key1].Version)
	assert.Equal(t, "2.0.0", byKey[key2].Version)
}

func TestGetServiceDetail_ReturnsLanguageAndMethods(t *testing.T) {
	consoleCtx, _, serviceStore := newTestConsoleContext(t)
	const mesh = "test-mesh"

	svcRes := newServiceResource(mesh,
		"org.apache.dubbo.samples.UserService", "1.0.0", "gray",
		[]string{"getUserById", "listUsers"})
	require.NoError(t, serviceStore.Add(svcRes))

	resp, err := GetServiceDetail(consoleCtx, model.BaseServiceReq{
		ServiceKeyValue: model.BuildServiceKey("org.apache.dubbo.samples.UserService", "1.0.0", "gray"),
		Mesh:            mesh,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "", resp.Language)
	assert.Equal(t, []string{"getUserById", "listUsers"}, resp.Methods)
}

func TestGetAppServiceInfo_ProvideSideQueriesServiceProjection(t *testing.T) {
	consoleCtx, _, serviceStore := newTestConsoleContext(t)
	const mesh = "test-mesh"

	require.NoError(t, serviceStore.Add(newServiceResourceForApp(mesh,
		"shopping-cart", "org.apache.dubbo.samples.UserService", "1.0.0", "gray",
		[]string{"getUserById"})))
	require.NoError(t, serviceStore.Add(newServiceResourceForApp(mesh,
		"order-center", "org.apache.dubbo.samples.OrderService", "1.0.0", "gray",
		[]string{"createOrder"})))

	resp, err := GetAppServiceInfo(consoleCtx, &model.ApplicationServiceFormReq{
		AppName: "shopping-cart",
		Side:    constants.ProviderSide,
		Mesh:    mesh,
		PageReq: coremodel.PageReq{PageOffset: 0, PageSize: 10},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	list, ok := resp.List.([]*model.ServiceSearchResp)
	require.True(t, ok)
	require.Len(t, list, 1)
	assert.Equal(t, "org.apache.dubbo.samples.UserService", list[0].ServiceName)
	assert.Equal(t, model.BuildServiceKey("org.apache.dubbo.samples.UserService", "1.0.0", "gray"), list[0].ServiceKey)
}
