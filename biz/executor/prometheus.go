package executor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.mokaz111.com/candy-agent/biz/model"
	"github.mokaz111.com/candy-agent/conf"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	pmodel "github.com/prometheus/common/model"
)

// PrometheusExecutor Prometheus 执行器
type PrometheusExecutor struct {
	client api.Client
	api    v1.API
	config conf.PrometheusConfig
}

// NewPrometheusExecutor 创建 Prometheus 执行器
func NewPrometheusExecutor(config conf.PrometheusConfig) (*PrometheusExecutor, error) {
	client, err := api.NewClient(api.Config{
		Address: config.URL,
	})
	if err != nil {
		hlog.Errorf("Failed to create Prometheus client: %v", err)
		return nil, err
	}

	return &PrometheusExecutor{
		client: client,
		api:    v1.NewAPI(client),
		config: config,
	}, nil
}

// Name 执行器名称
func (e *PrometheusExecutor) Name() string {
	return "prometheus"
}

// Execute 执行巡检项
func (e *PrometheusExecutor) Execute(ctx context.Context, item model.TaskItem) (model.TaskResult, error) {
	startTime := time.Now()
	result := model.TaskResult{
		ItemID: item.ID,
		Status: model.ResultStatusNormal,
	}

	// 获取查询参数
	query, ok := item.Params["query"].(string)
	if !ok {
		result.Status = model.ResultStatusFailed
		result.Message = "Missing query parameter"
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("missing query parameter")
	}

	// 设置超时上下文
	timeout := e.config.Timeout
	if timeout <= 0 {
		timeout = 10
	}
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 执行查询
	queryResult, warnings, err := e.api.Query(queryCtx, query, time.Now())
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Query failed: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// 处理警告
	if len(warnings) > 0 {
		hlog.Warnf("Prometheus query warnings: %v", warnings)
	}

	// 解析结果
	value, allValues, err := parsePrometheusResult(queryResult)
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to parse result: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}
	hlog.Infof("Query result: %s", value)

	// 检查阈值
	if thresholdStr, ok := item.Params["threshold"].(string); ok {
		threshold, err := strconv.ParseFloat(thresholdStr, 64)
		if err == nil {
			// 检查是否为向量结果
			if queryResult.Type() == pmodel.ValVector && len(allValues) > 0 {
				// 记录超过阈值的节点
				var exceedingNodes []string
				var highestValue float64
				var highestInstance string

				for instance, valueStr := range allValues {
					valueFloat, err := strconv.ParseFloat(valueStr, 64)
					if err == nil {
						if valueFloat > threshold {
							exceedingNodes = append(exceedingNodes, fmt.Sprintf("%s (%.2f)", instance, valueFloat))
							if valueFloat > highestValue {
								highestValue = valueFloat
								highestInstance = instance
							}
						}
					}
				}

				// 如果有节点超过阈值
				if len(exceedingNodes) > 0 {
					result.Status = model.ResultStatusWarning
					if len(exceedingNodes) == 1 {
						result.Message = fmt.Sprintf("节点 %s 的值 %.2f 超过阈值 %.2f", highestInstance, highestValue, threshold)
					} else {
						result.Message = fmt.Sprintf("有 %d 个节点超过阈值 %.2f，最高值为 %s 的 %.2f", len(exceedingNodes), threshold, highestInstance, highestValue)
					}
					// 在详情中添加所有超过阈值的节点
					result.Details = fmt.Sprintf("Query: %s\nResult: %s\n\n超过阈值的节点：\n%s", query, queryResult.String(), strings.Join(exceedingNodes, "\n"))
					result.Duration = time.Since(startTime).Milliseconds()
					return result, nil
				}

				result.Message = fmt.Sprintf("所有节点的值都在阈值 %.2f 范围内", threshold)
			} else {
				// 对于非向量结果，使用单一值比较
				floatValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					hlog.Warnf("Failed to parse value as float: %v", err)
				} else {
					// 根据阈值判断状态
					if floatValue > threshold {
						result.Status = model.ResultStatusWarning
						result.Message = fmt.Sprintf("值 %.2f 超过阈值 %.2f", floatValue, threshold)
					} else {
						result.Message = fmt.Sprintf("值 %.2f 在阈值 %.2f 范围内", floatValue, threshold)
					}
				}
			}
		}
	}

	// 设置详细信息
	result.Details = fmt.Sprintf("Query: %s\nResult: %s", query, queryResult.String())
	result.Duration = time.Since(startTime).Milliseconds()

	return result, nil
}

// parsePrometheusResult 解析 Prometheus 查询结果
func parsePrometheusResult(value pmodel.Value) (string, map[string]string, error) {
	allValues := make(map[string]string)

	switch value.Type() {
	case pmodel.ValScalar:
		scalar := value.(*pmodel.Scalar)
		valueStr := scalar.Value.String()
		allValues["scalar"] = valueStr
		return valueStr, allValues, nil

	case pmodel.ValVector:
		vector := value.(pmodel.Vector)
		if len(vector) == 0 {
			return "0", allValues, nil
		}

		// 收集所有样本的值
		for _, sample := range vector {
			// 提取实例名称或使用其他标签作为标识
			instance := string(sample.Metric["instance"])
			if instance == "" {
				// 如果没有instance标签，使用所有标签作为标识
				parts := make([]string, 0, len(sample.Metric))
				for k, v := range sample.Metric {
					parts = append(parts, fmt.Sprintf("%s=%s", k, v))
				}
				if len(parts) > 0 {
					instance = strings.Join(parts, ",")
				} else {
					instance = "unknown"
				}
			}

			// 保存值
			allValues[instance] = sample.Value.String()
		}

		// 返回第一个结果的值用于向后兼容
		return vector[0].Value.String(), allValues, nil

	case pmodel.ValMatrix:
		matrix := value.(pmodel.Matrix)
		if len(matrix) == 0 {
			return "0", allValues, nil
		}

		// 收集所有时间序列的最新值
		for _, series := range matrix {
			if len(series.Values) == 0 {
				continue
			}

			// 提取实例名称或使用其他标签作为标识
			instance := string(series.Metric["instance"])
			if instance == "" {
				// 如果没有instance标签，使用所有标签作为标识
				parts := make([]string, 0, len(series.Metric))
				for k, v := range series.Metric {
					parts = append(parts, fmt.Sprintf("%s=%s", k, v))
				}
				if len(parts) > 0 {
					instance = strings.Join(parts, ",")
				} else {
					instance = "unknown"
				}
			}

			// 使用最新的值
			lastPoint := series.Values[len(series.Values)-1]
			allValues[instance] = lastPoint.Value.String()
		}

		// 返回第一个时间序列的最后一个值用于向后兼容
		series := matrix[0]
		if len(series.Values) == 0 {
			return "0", allValues, nil
		}
		return series.Values[len(series.Values)-1].Value.String(), allValues, nil

	case pmodel.ValString:
		str := value.(*pmodel.String)
		allValues["string"] = str.Value
		return str.Value, allValues, nil

	default:
		return "", allValues, fmt.Errorf("unsupported value type: %s", value.Type().String())
	}
}
