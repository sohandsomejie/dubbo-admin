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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	meshproto "github.com/apache/dubbo-admin/api/mesh/v1alpha1"
	"github.com/apache/dubbo-admin/pkg/common/constants"
	"github.com/apache/dubbo-admin/pkg/config/app"
	discoverycfg "github.com/apache/dubbo-admin/pkg/config/discovery"
	consolectx "github.com/apache/dubbo-admin/pkg/console/context"
	"github.com/apache/dubbo-admin/pkg/core/governor"
	"github.com/apache/dubbo-admin/pkg/core/manager"
	meshresource "github.com/apache/dubbo-admin/pkg/core/resource/apis/mesh/v1alpha1"
	coremodel "github.com/apache/dubbo-admin/pkg/core/resource/model"
	corestore "github.com/apache/dubbo-admin/pkg/core/store"
	memorystore "github.com/apache/dubbo-admin/pkg/store/memory"
)

type testRuleGovernor struct {
	store corestore.ResourceStore
}

func (g *testRuleGovernor) CreateRule(r coremodel.Resource) error {
	return g.store.Add(r)
}

func (g *testRuleGovernor) UpdateRule(r coremodel.Resource) error {
	return g.store.Update(r)
}

func (g *testRuleGovernor) DeleteRule(r coremodel.Resource) error {
	return g.store.Delete(r)
}

type errorRuleGovernor struct {
	createErr error
	updateErr error
	deleteErr error
}

func (g *errorRuleGovernor) CreateRule(_ coremodel.Resource) error {
	return g.createErr
}

func (g *errorRuleGovernor) UpdateRule(_ coremodel.Resource) error {
	return g.updateErr
}

func (g *errorRuleGovernor) DeleteRule(_ coremodel.Resource) error {
	return g.deleteErr
}

type testGovernorRouter struct {
	governorsByKind map[coremodel.ResourceKind]governor.RuleGovernor
	defaultGovernor governor.RuleGovernor
}

func (r *testGovernorRouter) ResourceRoute(res coremodel.Resource) (governor.RuleGovernor, error) {
	gov, ok := r.governorsByKind[res.ResourceKind()]
	if !ok {
		return nil, fmt.Errorf("governor for kind %s not found", res.ResourceKind())
	}
	return gov, nil
}

func (r *testGovernorRouter) ResourceMeshRoute(_ string) (governor.RuleGovernor, error) {
	if r.defaultGovernor == nil {
		return nil, fmt.Errorf("default governor is nil")
	}
	return r.defaultGovernor, nil
}

func newInstanceServiceTestContext(
	t *testing.T) (consolectx.Context, corestore.ResourceStore, corestore.ResourceStore) {
	t.Helper()

	conditionStore := memorystore.NewMemoryResourceStore(meshresource.ConditionRouteKind)
	require.NoError(t, conditionStore.Init(nil))

	configuratorStore := memorystore.NewMemoryResourceStore(meshresource.DynamicConfigKind)
	require.NoError(t, configuratorStore.Init(nil))

	governorRouter := &testGovernorRouter{
		governorsByKind: map[coremodel.ResourceKind]governor.RuleGovernor{
			meshresource.ConditionRouteKind: &testRuleGovernor{store: conditionStore},
			meshresource.DynamicConfigKind:  &testRuleGovernor{store: configuratorStore},
		},
		defaultGovernor: &testRuleGovernor{store: conditionStore},
	}

	rm := manager.NewResourceManager(&testStoreRouter{
		stores: map[coremodel.ResourceKind]corestore.ResourceStore{
			meshresource.ConditionRouteKind: conditionStore,
			meshresource.DynamicConfigKind:  configuratorStore,
		},
	}, governorRouter)

	cfg := app.DefaultAdminConfig()
	cfg.Engine.Name = "test-engine"
	cfg.Discovery = []*discoverycfg.Config{
		{ID: "test-mesh", Name: "test-registry"},
	}

	return &testConsoleContext{
		rm:  rm,
		cfg: cfg,
	}, conditionStore, configuratorStore
}

func newInstanceServiceTestContextWithStores(
	t *testing.T,
	stores map[coremodel.ResourceKind]corestore.ResourceStore) consolectx.Context {
	t.Helper()
	return newInstanceServiceTestContextWithStoresAndGovernors(t, stores, nil)
}

func newInstanceServiceTestContextWithStoresAndGovernors(
	t *testing.T,
	stores map[coremodel.ResourceKind]corestore.ResourceStore,
	overrideGovernors map[coremodel.ResourceKind]governor.RuleGovernor) consolectx.Context {
	t.Helper()

	governorsByKind := make(map[coremodel.ResourceKind]governor.RuleGovernor, len(stores))
	for kind, store := range stores {
		gov := governor.RuleGovernor(&testRuleGovernor{store: store})
		if overrideGovernors != nil {
			if override, ok := overrideGovernors[kind]; ok {
				gov = override
			}
		}
		governorsByKind[kind] = gov
	}
	governorRouter := &testGovernorRouter{
		governorsByKind: governorsByKind,
	}

	rm := manager.NewResourceManager(&testStoreRouter{
		stores: stores,
	}, governorRouter)

	cfg := app.DefaultAdminConfig()
	cfg.Engine.Name = "test-engine"
	cfg.Discovery = []*discoverycfg.Config{
		{ID: "test-mesh", Name: "test-registry"},
	}

	return &testConsoleContext{
		rm:  rm,
		cfg: cfg,
	}
}

func newConditionRouteResource(mesh, appName string, conditions []string) *meshresource.ConditionRouteResource {
	ruleName := appName + constants.ConditionRuleDotSuffix
	res := meshresource.NewConditionRouteResourceWithAttributes(ruleName, mesh)
	res.Spec = &meshproto.ConditionRoute{
		ConfigVersion: constants.ConfiguratorVersionV3,
		Enabled:       true,
		Force:         true,
		Runtime:       true,
		Key:           appName,
		Scope:         constants.ScopeApplication,
		Conditions:    conditions,
	}
	return res
}

func getConditionRouteResource(
	t *testing.T,
	store corestore.ResourceStore,
	mesh string,
	appName string) *meshresource.ConditionRouteResource {
	t.Helper()
	ruleName := appName + constants.ConditionRuleDotSuffix
	raw, exists, err := store.GetByKey(coremodel.BuildResourceKey(mesh, ruleName))
	require.NoError(t, err)
	require.True(t, exists)
	res, ok := raw.(*meshresource.ConditionRouteResource)
	require.True(t, ok)
	return res
}

func newDynamicConfigResource(mesh, appName string, configs []*meshproto.OverrideConfig) *meshresource.DynamicConfigResource {
	ruleName := appName + constants.ConfiguratorRuleDotSuffix
	res := meshresource.NewDynamicConfigResourceWithAttributes(ruleName, mesh)
	res.Spec = &meshproto.DynamicConfig{
		Key:           appName,
		Scope:         constants.ScopeApplication,
		ConfigVersion: constants.ConfiguratorVersionV3,
		Enabled:       true,
		Configs:       configs,
	}
	return res
}

func getDynamicConfigResource(
	t *testing.T,
	store corestore.ResourceStore,
	mesh string,
	appName string) *meshresource.DynamicConfigResource {
	t.Helper()
	ruleName := appName + constants.ConfiguratorRuleDotSuffix
	raw, exists, err := store.GetByKey(coremodel.BuildResourceKey(mesh, ruleName))
	require.NoError(t, err)
	require.True(t, exists)
	res, ok := raw.(*meshresource.DynamicConfigResource)
	require.True(t, ok)
	return res
}

func assertConditionRouteNotExists(t *testing.T, store corestore.ResourceStore, mesh string, appName string) {
	t.Helper()
	ruleName := appName + constants.ConditionRuleDotSuffix
	_, exists, err := store.GetByKey(coremodel.BuildResourceKey(mesh, ruleName))
	require.NoError(t, err)
	assert.False(t, exists)
}

func assertDynamicConfigNotExists(t *testing.T, store corestore.ResourceStore, mesh string, appName string) {
	t.Helper()
	ruleName := appName + constants.ConfiguratorRuleDotSuffix
	_, exists, err := store.GetByKey(coremodel.BuildResourceKey(mesh, ruleName))
	require.NoError(t, err)
	assert.False(t, exists)
}

func openAccessLogConfig(ip string) *meshproto.OverrideConfig {
	return &meshproto.OverrideConfig{
		Side: constants.SideProvider,
		Match: &meshproto.ConditionMatch{
			Address: &meshproto.AddressMatch{Wildcard: ip + `:*`},
		},
		Parameters:    map[string]string{`accesslog`: `true`},
		XGenerateByCp: true,
	}
}

func TestUpdateInstanceTrafficStatus_CreateRuleWhenDisableOnAndRuleMissing(t *testing.T) {
	ctx, conditionStore, _ := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	require.NoError(t, UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, true))

	res := getConditionRouteResource(t, conditionStore, mesh, appName)
	assert.Contains(t, res.Spec.Conditions, disableExpression(instanceIP))
}

func TestUpdateInstanceTrafficStatus_NoRuleWhenEnableOnAndRuleMissing(t *testing.T) {
	ctx, conditionStore, _ := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	require.NoError(t, UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, false))

	assertConditionRouteNotExists(t, conditionStore, mesh, appName)
}

func TestUpdateInstanceTrafficStatus_RemoveRuleWhenEnableOnDisabledInstance(t *testing.T) {
	ctx, conditionStore, _ := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	res := newConditionRouteResource(mesh, appName, []string{
		"consumer.host=10.10.10.10",
		disableExpression(instanceIP),
	})
	require.NoError(t, conditionStore.Add(res))

	require.NoError(t, UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, false))

	updated := getConditionRouteResource(t, conditionStore, mesh, appName)
	assert.NotContains(t, updated.Spec.Conditions, disableExpression(instanceIP))
	assert.Contains(t, updated.Spec.Conditions, "consumer.host=10.10.10.10")
}

func TestUpdateInstanceTrafficStatus_AppendDisableRuleWhenDisableOnEnabledInstance(t *testing.T) {
	ctx, conditionStore, _ := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	res := newConditionRouteResource(mesh, appName, []string{"consumer.host=10.10.10.10"})
	require.NoError(t, conditionStore.Add(res))

	require.NoError(t, UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, true))

	updated := getConditionRouteResource(t, conditionStore, mesh, appName)
	assert.Contains(t, updated.Spec.Conditions, "consumer.host=10.10.10.10")
	assert.Contains(t, updated.Spec.Conditions, disableExpression(instanceIP))
}

func TestUpdateInstanceTrafficStatus_ReturnErrorWhenConditionStoreMissing(t *testing.T) {
	configuratorStore := memorystore.NewMemoryResourceStore(meshresource.DynamicConfigKind)
	require.NoError(t, configuratorStore.Init(nil))
	ctx := newInstanceServiceTestContextWithStores(t, map[coremodel.ResourceKind]corestore.ResourceStore{
		meshresource.DynamicConfigKind: configuratorStore,
	})
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	err := UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, true)
	require.Error(t, err)
}

func TestUpdateInstanceTrafficStatus_NoopWhenDisableOnAlreadyDisabledInstance(t *testing.T) {
	ctx, conditionStore, _ := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	res := newConditionRouteResource(mesh, appName, []string{
		"consumer.host=10.10.10.10",
		disableExpression(instanceIP),
	})
	require.NoError(t, conditionStore.Add(res))

	require.NoError(t, UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, true))

	updated := getConditionRouteResource(t, conditionStore, mesh, appName)
	assert.Equal(t, []string{
		"consumer.host=10.10.10.10",
		disableExpression(instanceIP),
	}, updated.Spec.Conditions)
}

func TestUpdateInstanceTrafficStatus_NoopWhenEnableOnAlreadyEnabledInstance(t *testing.T) {
	ctx, conditionStore, _ := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	res := newConditionRouteResource(mesh, appName, []string{"consumer.host=10.10.10.10"})
	require.NoError(t, conditionStore.Add(res))

	require.NoError(t, UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, false))

	updated := getConditionRouteResource(t, conditionStore, mesh, appName)
	assert.Equal(t, []string{"consumer.host=10.10.10.10"}, updated.Spec.Conditions)
}

func TestUpdateInstanceTrafficStatus_ReturnErrorWhenCreateConditionRuleFailed(t *testing.T) {
	conditionStore := memorystore.NewMemoryResourceStore(meshresource.ConditionRouteKind)
	require.NoError(t, conditionStore.Init(nil))
	ctx := newInstanceServiceTestContextWithStoresAndGovernors(
		t,
		map[coremodel.ResourceKind]corestore.ResourceStore{
			meshresource.ConditionRouteKind: conditionStore,
		},
		map[coremodel.ResourceKind]governor.RuleGovernor{
			meshresource.ConditionRouteKind: &errorRuleGovernor{createErr: fmt.Errorf("create failed")},
		},
	)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	err := UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, true)
	require.Error(t, err)
}

func TestUpdateInstanceTrafficStatus_ReturnErrorWhenUpdateConditionRuleFailed(t *testing.T) {
	conditionStore := memorystore.NewMemoryResourceStore(meshresource.ConditionRouteKind)
	require.NoError(t, conditionStore.Init(nil))
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"
	res := newConditionRouteResource(mesh, appName, []string{disableExpression(instanceIP)})
	require.NoError(t, conditionStore.Add(res))
	ctx := newInstanceServiceTestContextWithStoresAndGovernors(
		t,
		map[coremodel.ResourceKind]corestore.ResourceStore{
			meshresource.ConditionRouteKind: conditionStore,
		},
		map[coremodel.ResourceKind]governor.RuleGovernor{
			meshresource.ConditionRouteKind: &errorRuleGovernor{updateErr: fmt.Errorf("update failed")},
		},
	)

	err := UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, false)
	require.Error(t, err)
}

func TestUpdateInstanceTrafficStatus_ReturnErrorWhenAppendDisableRuleUpdateFailed(t *testing.T) {
	conditionStore := memorystore.NewMemoryResourceStore(meshresource.ConditionRouteKind)
	require.NoError(t, conditionStore.Init(nil))
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"
	res := newConditionRouteResource(mesh, appName, []string{"consumer.host=10.10.10.10"})
	require.NoError(t, conditionStore.Add(res))
	ctx := newInstanceServiceTestContextWithStoresAndGovernors(
		t,
		map[coremodel.ResourceKind]corestore.ResourceStore{
			meshresource.ConditionRouteKind: conditionStore,
		},
		map[coremodel.ResourceKind]governor.RuleGovernor{
			meshresource.ConditionRouteKind: &errorRuleGovernor{updateErr: fmt.Errorf("update failed")},
		},
	)

	err := UpdateInstanceTrafficStatus(ctx, mesh, appName, instanceIP, true)
	require.Error(t, err)
}

func TestUpdateInstanceAccessLogOpenStatus_CreateConfigWhenOpenAndConfigMissing(t *testing.T) {
	ctx, _, configuratorStore := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	require.NoError(t, UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, true))

	res := getDynamicConfigResource(t, configuratorStore, mesh, appName)
	require.Len(t, res.Spec.Configs, 1)
	assert.True(t, isInstanceAccessLogOpen(res.Spec.Configs[0], instanceIP))
}

func TestUpdateInstanceAccessLogOpenStatus_NoConfigWhenCloseAndConfigMissing(t *testing.T) {
	ctx, _, configuratorStore := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	require.NoError(t, UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, false))

	assertDynamicConfigNotExists(t, configuratorStore, mesh, appName)
}

func TestUpdateInstanceAccessLogOpenStatus_ReturnErrorWhenConfiguratorStoreMissing(t *testing.T) {
	conditionStore := memorystore.NewMemoryResourceStore(meshresource.ConditionRouteKind)
	require.NoError(t, conditionStore.Init(nil))
	ctx := newInstanceServiceTestContextWithStores(t, map[coremodel.ResourceKind]corestore.ResourceStore{
		meshresource.ConditionRouteKind: conditionStore,
	})
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	err := UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, true)
	require.Error(t, err)
}

func TestUpdateInstanceAccessLogOpenStatus_RemoveConfigWhenCloseAndOpenConfigExists(t *testing.T) {
	ctx, _, configuratorStore := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	res := newDynamicConfigResource(mesh, appName, []*meshproto.OverrideConfig{
		openAccessLogConfig(instanceIP),
		openAccessLogConfig("10.0.0.2"),
	})
	require.NoError(t, configuratorStore.Add(res))

	require.NoError(t, UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, false))

	updated := getDynamicConfigResource(t, configuratorStore, mesh, appName)
	assert.Len(t, updated.Spec.Configs, 1)
	assert.True(t, isInstanceAccessLogOpen(updated.Spec.Configs[0], "10.0.0.2"))
	assert.False(t, isInstanceAccessLogOpen(updated.Spec.Configs[0], instanceIP))
}

func TestUpdateInstanceAccessLogOpenStatus_AppendConfigWhenOpenAndNoMatchedConfig(t *testing.T) {
	ctx, _, configuratorStore := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	res := newDynamicConfigResource(mesh, appName, []*meshproto.OverrideConfig{
		openAccessLogConfig("10.0.0.2"),
	})
	require.NoError(t, configuratorStore.Add(res))

	require.NoError(t, UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, true))

	updated := getDynamicConfigResource(t, configuratorStore, mesh, appName)
	assert.Len(t, updated.Spec.Configs, 2)
	assert.True(t, isInstanceAccessLogOpen(updated.Spec.Configs[0], "10.0.0.2"))
	assert.True(t, isInstanceAccessLogOpen(updated.Spec.Configs[1], instanceIP))
}

func TestUpdateInstanceAccessLogOpenStatus_NoopWhenOpenAndConfigAlreadyExists(t *testing.T) {
	ctx, _, configuratorStore := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	res := newDynamicConfigResource(mesh, appName, []*meshproto.OverrideConfig{
		openAccessLogConfig(instanceIP),
		openAccessLogConfig("10.0.0.2"),
	})
	require.NoError(t, configuratorStore.Add(res))

	require.NoError(t, UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, true))

	updated := getDynamicConfigResource(t, configuratorStore, mesh, appName)
	assert.Len(t, updated.Spec.Configs, 2)
	assert.True(t, isInstanceAccessLogOpen(updated.Spec.Configs[0], instanceIP))
	assert.True(t, isInstanceAccessLogOpen(updated.Spec.Configs[1], "10.0.0.2"))
}

func TestUpdateInstanceAccessLogOpenStatus_NoopWhenCloseAndNoMatchedConfig(t *testing.T) {
	ctx, _, configuratorStore := newInstanceServiceTestContext(t)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	res := newDynamicConfigResource(mesh, appName, []*meshproto.OverrideConfig{
		openAccessLogConfig("10.0.0.2"),
	})
	require.NoError(t, configuratorStore.Add(res))

	require.NoError(t, UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, false))

	updated := getDynamicConfigResource(t, configuratorStore, mesh, appName)
	assert.Len(t, updated.Spec.Configs, 1)
	assert.True(t, isInstanceAccessLogOpen(updated.Spec.Configs[0], "10.0.0.2"))
}

func TestUpdateInstanceAccessLogOpenStatus_ReturnErrorWhenCreateConfiguratorFailed(t *testing.T) {
	configuratorStore := memorystore.NewMemoryResourceStore(meshresource.DynamicConfigKind)
	require.NoError(t, configuratorStore.Init(nil))
	ctx := newInstanceServiceTestContextWithStoresAndGovernors(
		t,
		map[coremodel.ResourceKind]corestore.ResourceStore{
			meshresource.DynamicConfigKind: configuratorStore,
		},
		map[coremodel.ResourceKind]governor.RuleGovernor{
			meshresource.DynamicConfigKind: &errorRuleGovernor{createErr: fmt.Errorf("create failed")},
		},
	)
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"

	err := UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, true)
	require.Error(t, err)
}

func TestUpdateInstanceAccessLogOpenStatus_ReturnErrorWhenUpdateConfiguratorFailed(t *testing.T) {
	configuratorStore := memorystore.NewMemoryResourceStore(meshresource.DynamicConfigKind)
	require.NoError(t, configuratorStore.Init(nil))
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"
	res := newDynamicConfigResource(mesh, appName, []*meshproto.OverrideConfig{
		openAccessLogConfig(instanceIP),
	})
	require.NoError(t, configuratorStore.Add(res))
	ctx := newInstanceServiceTestContextWithStoresAndGovernors(
		t,
		map[coremodel.ResourceKind]corestore.ResourceStore{
			meshresource.DynamicConfigKind: configuratorStore,
		},
		map[coremodel.ResourceKind]governor.RuleGovernor{
			meshresource.DynamicConfigKind: &errorRuleGovernor{updateErr: fmt.Errorf("update failed")},
		},
	)

	err := UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, false)
	require.Error(t, err)
}

func TestUpdateInstanceAccessLogOpenStatus_ReturnErrorWhenAppendConfigUpdateFailed(t *testing.T) {
	configuratorStore := memorystore.NewMemoryResourceStore(meshresource.DynamicConfigKind)
	require.NoError(t, configuratorStore.Init(nil))
	const mesh, appName, instanceIP = "test-mesh", "shop-comment", "10.0.0.1"
	res := newDynamicConfigResource(mesh, appName, []*meshproto.OverrideConfig{
		openAccessLogConfig("10.0.0.2"),
	})
	require.NoError(t, configuratorStore.Add(res))
	ctx := newInstanceServiceTestContextWithStoresAndGovernors(
		t,
		map[coremodel.ResourceKind]corestore.ResourceStore{
			meshresource.DynamicConfigKind: configuratorStore,
		},
		map[coremodel.ResourceKind]governor.RuleGovernor{
			meshresource.DynamicConfigKind: &errorRuleGovernor{updateErr: fmt.Errorf("update failed")},
		},
	)

	err := UpdateInstanceAccessLogOpenStatus(ctx, mesh, appName, instanceIP, true)
	require.Error(t, err)
}
