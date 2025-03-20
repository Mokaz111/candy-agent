package utils

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseYAMLToGroups 将YAML格式的告警规则解析为规则组
func ParseYAMLToGroups(yamlContent string) ([]map[string]interface{}, error) {
	// 解析YAML内容
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	// 提取规则组
	groups, ok := data["groups"]
	if !ok {
		return nil, fmt.Errorf("no groups found in YAML content")
	}

	// 转换为规则组列表
	groupsList, ok := groups.([]interface{})
	if !ok {
		return nil, fmt.Errorf("groups is not a list")
	}

	// 转换为所需的格式
	result := make([]map[string]interface{}, 0, len(groupsList))
	for _, g := range groupsList {
		group, ok := g.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("group is not a map")
		}
		result = append(result, group)
	}

	return result, nil
}

// ParsePrometheusRules 解析Prometheus规则YAML
func ParsePrometheusRules(yamlContent string) ([]map[string]interface{}, error) {
	return ParseYAMLToGroups(yamlContent)
}

// ParseVMRules 解析VictoriaMetrics规则YAML
func ParseVMRules(yamlContent string) ([]map[string]interface{}, error) {
	return ParseYAMLToGroups(yamlContent)
}
