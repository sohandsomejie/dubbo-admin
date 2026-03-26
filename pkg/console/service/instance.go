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
	"io"
	"net/http"
	"strings"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"

	meshproto "github.com/apache/dubbo-admin/api/mesh/v1alpha1"
	"github.com/apache/dubbo-admin/pkg/common/bizerror"
	"github.com/apache/dubbo-admin/pkg/common/constants"
	consolectx "github.com/apache/dubbo-admin/pkg/console/context"
	"github.com/apache/dubbo-admin/pkg/console/model"
	"github.com/apache/dubbo-admin/pkg/core/logger"
	"github.com/apache/dubbo-admin/pkg/core/manager"
	meshresource "github.com/apache/dubbo-admin/pkg/core/resource/apis/mesh/v1alpha1"
	coremodel "github.com/apache/dubbo-admin/pkg/core/resource/model"
	"github.com/apache/dubbo-admin/pkg/core/store/index"
)

// SearchInstanceByIp search instance by ip
func SearchInstanceByIp(ctx consolectx.Context, req *model.SearchReq) (*model.SearchPaginationResult, error) {
	pageData, err := manager.PageListByIndexes[*meshresource.InstanceResource](
		ctx.ResourceManager(),
		meshresource.InstanceKind,
		map[string]string{
			index.ByMeshIndex:       req.Mesh,
			index.ByInstanceIpIndex: req.Keywords,
		},
		req.PageReq)
	if err != nil {
		return nil, err
	}
	if pageData.Data == nil || len(pageData.Data) == 0 {
		return &model.SearchPaginationResult{
			List: []*meshresource.ServiceProviderMetadataResourceList{},
			PageInfo: coremodel.Pagination{
				Total:      0,
				PageSize:   req.PageReq.PageSize,
				PageOffset: req.PageReq.PageOffset,
			},
		}, nil
	}
	return &model.SearchPaginationResult{
		List: slice.Map(pageData.Data, func(_ int, item *meshresource.InstanceResource) *model.AppInstanceInfoResp {
			return buildAppInstanceInfoResp(item, ctx.Config())
		}),
		PageInfo: pageData.Pagination,
	}, nil
}

// SearchInstanceByName search instance by name
func SearchInstanceByName(ctx consolectx.Context, req *model.SearchReq) (*model.SearchPaginationResult, error) {
	pageData, err := manager.PageListByIndexes[*meshresource.InstanceResource](
		ctx.ResourceManager(),
		meshresource.InstanceKind,
		map[string]string{
			index.ByMeshIndex:         req.Mesh,
			index.ByInstanceNameIndex: req.Keywords,
		},
		req.PageReq)
	if err != nil {
		return nil, err
	}
	if pageData.Data == nil || len(pageData.Data) == 0 {
		return &model.SearchPaginationResult{
			List: []*meshresource.ServiceProviderMetadataResourceList{},
			PageInfo: coremodel.Pagination{
				Total:      0,
				PageSize:   req.PageReq.PageSize,
				PageOffset: req.PageReq.PageOffset,
			},
		}, nil
	}
	return &model.SearchPaginationResult{
		List: slice.Map(pageData.Data, func(_ int, item *meshresource.InstanceResource) *model.AppInstanceInfoResp {
			return buildAppInstanceInfoResp(item, ctx.Config())
		}),
		PageInfo: pageData.Pagination,
	}, nil
}

func SearchInstances(ctx consolectx.Context, req *model.SearchInstanceReq) (*model.SearchPaginationResult, error) {
	if strutil.IsNotBlank(req.Keywords) {
		return SearchInstanceByIp(ctx, &model.SearchReq{
			PageReq:    req.PageReq,
			SearchType: "ip",
			Keywords:   req.Keywords,
			Mesh:       req.Mesh,
		})
	}
	pageData, err := manager.PageListByIndexes[*meshresource.InstanceResource](
		ctx.ResourceManager(),
		meshresource.InstanceKind,
		map[string]string{
			index.ByMeshIndex: req.Mesh,
		},
		req.PageReq)
	if err != nil {
		logger.Errorf("Failed to search instance,req: %s, cause: %v", convertor.ToString(req), err)
		return nil, err
	}
	resp := model.NewSearchPaginationResult()
	var list []*model.SearchInstanceResp
	for _, item := range pageData.Data {
		list = append(list, model.NewSearchInstanceResp().FromInstanceResource(item, ctx.Config()))
	}
	resp.List = list
	resp.PageInfo = pageData.Pagination
	return resp, nil
}

func GetInstanceDetail(ctx consolectx.Context, req *model.InstanceDetailReq) (*model.InstanceDetailResp, error) {
	res, _, err := manager.GetByKey[*meshresource.InstanceResource](
		ctx.ResourceManager(),
		meshresource.InstanceKind,
		coremodel.BuildResourceKey(req.Mesh, req.InstanceName),
	)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, bizerror.New(bizerror.NotFoundError, fmt.Sprintf("instance %s not found", req.InstanceName))
	}

	resp := model.FromInstanceResource(res, ctx.Config())
	return resp, nil
}

func GetInstanceMetrics(ctx consolectx.Context, req *model.MetricsReq) ([]*model.MetricsResp, error) {
	res, _, err := manager.GetByKey[*meshresource.InstanceResource](
		ctx.ResourceManager(),
		meshresource.InstanceKind,
		coremodel.BuildResourceKey(req.Mesh, req.InstanceName),
	)
	if err != nil {
		return nil, err
	}
	instance := res.Spec
	metricsData, err := fetchMetricsData(instance.Ip, instance.QosPort)
	if err != nil {
		return nil, err
	}

	metricsResp := &model.MetricsResp{
		InstanceName: instance.Name,
		Metrics:      metricsData,
	}

	return []*model.MetricsResp{metricsResp}, nil
}

func UpdateInstanceTrafficStatus(ctx consolectx.Context, mesh string, appName string, instanceIP string, newDisabled bool) error {
	conditionRuleName := appName + constants.ConditionRuleDotSuffix
	conditionRuleRes, err := GetConditionRule(ctx, conditionRuleName, mesh)
	if err != nil {
		logger.Errorf("get condition rule for %s failed, cause: %s", appName, err)
		return err
	}
	if conditionRuleRes == nil || conditionRuleRes.Spec == nil {
		return createTrafficConditionRuleIfNeeded(ctx, mesh, appName, instanceIP, newDisabled)
	}
	return updateExistingTrafficConditionRule(ctx, conditionRuleRes, appName, instanceIP, newDisabled)
}

func createTrafficConditionRuleIfNeeded(
	ctx consolectx.Context,
	mesh string,
	appName string,
	instanceIP string,
	newDisabled bool) error {
	if !newDisabled {
		return nil
	}
	conditionRoute := generateDefaultConditionV3(true, true, true, appName, constants.ScopeApplication)
	conditionRoute.Conditions = append(conditionRoute.Conditions, disableExpression(instanceIP))
	resName := appName + constants.ConditionRuleDotSuffix
	conditionRuleRes := meshresource.NewConditionRouteResourceWithAttributes(resName, mesh)
	conditionRuleRes.Spec = conditionRoute
	if err := CreateConditionRule(ctx, conditionRuleRes); err != nil {
		logger.Errorf("create condition rule for app %s failed, cause: %s", appName, err)
		return err
	}
	return nil
}

func updateExistingTrafficConditionRule(
	ctx consolectx.Context,
	conditionRuleRes *meshresource.ConditionRouteResource,
	appName string,
	instanceIP string,
	newDisabled bool) error {
	disabledIndex, disabledFound := findDisabledTrafficConditionIndex(conditionRuleRes.Spec.Conditions, instanceIP)
	if disabledFound {
		if newDisabled {
			logger.Warnf("The instance %s has been disabled, skip updating condition rule", instanceIP)
			return nil
		}
		conditionRuleRes.Spec.Conditions = append(
			conditionRuleRes.Spec.Conditions[:disabledIndex],
			conditionRuleRes.Spec.Conditions[disabledIndex+1:]...)
		if err := UpdateConditionRule(ctx, conditionRuleRes); err != nil {
			logger.Errorf("update condition rule for app %s failed, cause: %s", appName, err)
			return err
		}
		return nil
	}
	if !newDisabled {
		logger.Warnf("the instance %s has been enabled, skip updating condition rule", instanceIP)
		return nil
	}
	conditionRuleRes.Spec.Conditions = append(conditionRuleRes.Spec.Conditions, disableExpression(instanceIP))
	if err := UpdateConditionRule(ctx, conditionRuleRes); err != nil {
		logger.Errorf("update condition rule for app %s failed, cause: %s", appName, err)
		return err
	}
	return nil
}

func findDisabledTrafficConditionIndex(conditions []string, instanceIP string) (int, bool) {
	for i, condition := range conditions {
		if isInstanceTrafficDisabled(condition, instanceIP) {
			return i, true
		}
	}
	return -1, false
}

func GetInstanceTrafficStatus(ctx consolectx.Context, mesh string, appName string, instanceIP string) (bool, error) {
	resName := appName + constants.ConditionRuleDotSuffix
	res, err := GetConditionRule(ctx, resName, mesh)
	if err != nil {
		logger.Errorf("get condition rule for %s failed, cause: %s", appName, err)
		return true, err
	}
	if res == nil {
		return false, nil
	}
	disabled := false
	slice.ForEachWithBreak(res.Spec.Conditions, func(_ int, condition string) bool {
		disabled = isInstanceTrafficDisabled(condition, instanceIP)
		return disabled
	})
	return disabled, nil
}

func disableExpression(instanceIP string) string {
	return "=>host!=" + instanceIP
}

func generateDefaultConditionV3(Enabled, Force, Runtime bool, Key, Scope string) *meshproto.ConditionRoute {
	return &meshproto.ConditionRoute{
		ConfigVersion: constants.ConfiguratorVersionV3,
		Priority:      0,
		Enabled:       true,
		Force:         Force,
		Runtime:       Runtime,
		Key:           Key,
		Scope:         Scope,
		Conditions:    make([]string, 0),
	}
}

// isInstanceTrafficDisabled judge if a condition is instance traffic disable expression or not.
// A condition include fromCondition and toCondition which is seperated by `=>`.
// return true if the instance traffic is disabled, otherwise return false.
func isInstanceTrafficDisabled(condition string, targetIP string) bool {
	if len(condition) == 0 {
		return false
	}
	condition = strings.ReplaceAll(condition, " ", "")
	// only accept string start with `=>`
	if !strings.HasPrefix(condition, "=>") {
		return false
	}
	toCondition := strings.TrimPrefix(condition, "=>")

	if !strings.Contains(toCondition, targetIP) {
		return false
	}
	targetExpression := "host!=" + targetIP
	if targetExpression != toCondition {
		return false
	}
	return true
}

func GetInstanceAccessLogOpenStatus(ctx consolectx.Context, mesh string, applicationName string, instanceIP string) (bool, error) {
	appConfiguratorName := applicationName + constants.ConfiguratorRuleDotSuffix
	res, err := GetConfigurator(ctx, appConfiguratorName, mesh)
	if err != nil {
		logger.Errorf("get configurator for %s failed, cause: %s", appConfiguratorName, err)
		return false, err
	}
	if res == nil || res.Spec == nil {
		return false, nil
	}
	openAccessLog := false
	if res.Spec.Enabled {
		slice.ForEachWithBreak(res.Spec.Configs, func(_ int, conf *meshproto.OverrideConfig) bool {
			openAccessLog = isInstanceAccessLogOpen(conf, instanceIP)
			return openAccessLog
		})
	}
	return openAccessLog, nil
}

func UpdateInstanceAccessLogOpenStatus(
	ctx consolectx.Context,
	mesh string,
	appName string,
	instanceIP string,
	openStatus bool) error {
	appConfiguratorName := appName + constants.ConfiguratorRuleDotSuffix
	res, err := GetConfigurator(ctx, appConfiguratorName, mesh)
	if err != nil {
		logger.Errorf("get configurator for %s failed, cause: %s", appConfiguratorName, err)
		return err
	}
	if res == nil || res.Spec == nil {
		return createAccessLogConfiguratorIfNeeded(ctx, mesh, appName, instanceIP, openStatus)
	}
	return updateExistingAccessLogConfigurator(ctx, res, appName, instanceIP, openStatus)
}

func createAccessLogConfiguratorIfNeeded(
	ctx consolectx.Context,
	mesh string,
	appName string,
	instanceIP string,
	openStatus bool) error {
	if !openStatus {
		logger.Warnf("the instance %s accesslog is disabled, skip updating configurator", instanceIP)
		return nil
	}
	res := meshresource.NewDynamicConfigResourceWithAttributes(appName+constants.ConfiguratorRuleDotSuffix, mesh)
	res.Spec = &meshproto.DynamicConfig{
		Key:           appName,
		Scope:         constants.ScopeApplication,
		ConfigVersion: constants.ConfiguratorVersionV3,
		Enabled:       true,
		Configs: []*meshproto.OverrideConfig{
			{
				Side:          constants.SideProvider,
				Match:         &meshproto.ConditionMatch{Address: &meshproto.AddressMatch{Wildcard: instanceIP + `:*`}},
				Parameters:    map[string]string{`accesslog`: `true`},
				XGenerateByCp: true,
			},
		},
	}
	if err := CreateConfigurator(ctx, res); err != nil {
		logger.Errorf("create configurator for instance %s%s failed, cause: %s", appName, instanceIP, err)
		return err
	}
	return nil
}

func updateExistingAccessLogConfigurator(
	ctx consolectx.Context,
	res *meshresource.DynamicConfigResource,
	appName string,
	instanceIP string,
	openStatus bool) error {
	configIndex, accessLogOpened := findAccessLogConfigIndex(res.Spec.Configs, instanceIP)
	if accessLogOpened {
		if openStatus {
			logger.Warnf("the instance %s accesslog is already opened, skip updating configurator", instanceIP)
			return nil
		}
		res.Spec.Configs = slice.Concat(res.Spec.Configs[:configIndex], res.Spec.Configs[configIndex+1:])
		if err := UpdateConfigurator(ctx, res); err != nil {
			logger.Errorf("update configurator for instance %s%s failed, cause: %s", appName, instanceIP, err)
			return err
		}
		return nil
	}
	if !openStatus {
		logger.Warnf("the instance %s accesslog is already disabled, skip updating configurator", instanceIP)
		return nil
	}
	res.Spec.Configs = append(res.Spec.Configs, &meshproto.OverrideConfig{
		Side:          constants.SideProvider,
		Match:         &meshproto.ConditionMatch{Address: &meshproto.AddressMatch{Wildcard: instanceIP + `:*`}},
		Parameters:    map[string]string{`accesslog`: `true`},
		XGenerateByCp: true,
	})
	if err := UpdateConfigurator(ctx, res); err != nil {
		logger.Errorf("update configurator for instance %s%s failed, cause: %s", appName, instanceIP, err)
		return err
	}
	return nil
}

func findAccessLogConfigIndex(configs []*meshproto.OverrideConfig, instanceIP string) (int, bool) {
	for i, config := range configs {
		if isInstanceAccessLogOpen(config, instanceIP) {
			return i, true
		}
	}
	return -1, false
}

func isInstanceAccessLogOpen(conf *meshproto.OverrideConfig, IP string) bool {
	if conf != nil &&
		conf.Match != nil &&
		conf.Match.Address != nil &&
		conf.Match.Address.Wildcard == IP+`:*` &&
		conf.Side == constants.SideProvider &&
		conf.Parameters != nil &&
		conf.Parameters[`accesslog`] == `true` {
		return true
	}
	return false
}
func fetchMetricsData(ip string, port int64) ([]model.Metric, error) {
	url := fmt.Sprintf("http://%s:%d/metrics", ip, port)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	metrics, err := parsePrometheusData(string(body))
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

// parsePrometheusData parses Prometheus text format data and converts it to a slice of Metrics.
func parsePrometheusData(data string) ([]model.Metric, error) {
	var metrics []model.Metric
	for line := range strings.SplitSeq(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue
		}

		metricPart := parts[0]
		valuePart := parts[1]

		// Extract the metric name and labels
		nameAndLabels := strings.SplitN(metricPart, "{", 2)
		if len(nameAndLabels) != 2 {
			continue
		}

		name := nameAndLabels[0]
		labelsPart := strings.TrimSuffix(nameAndLabels[1], "}")

		labels := make(map[string]string)
		for label := range strings.SplitSeq(labelsPart, ",") {
			if label == "" {
				continue
			}
			labelParts := strings.SplitN(label, "=", 2)
			if len(labelParts) == 2 {
				labels[labelParts[0]] = strings.Trim(labelParts[1], `"`)
			}
		}

		// Parse the value
		var value float64
		fmt.Sscanf(valuePart, "%f", &value)

		metrics = append(metrics, model.Metric{
			Name:   name,
			Labels: labels,
			Value:  value,
		})
	}

	return metrics, nil
}
