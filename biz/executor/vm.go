package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.mokaz111.com/candy-agent/biz/model"
	config "github.mokaz111.com/candy-agent/conf"
)

// VMExecutor VictoriaMetrics 执行器
type VMExecutor struct {
	client  *http.Client
	baseURL string
	config  config.VMConfig
}

// VMQueryResponse VictoriaMetrics 查询响应
type VMQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string          `json:"resultType"`
		Result     json.RawMessage `json:"result"`
	} `json:"data"`
}

// VMScalarResult 标量结果
type VMScalarResult []interface{} // [时间戳, 值]

// VMVectorResult 向量结果
type VMVectorResult []struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"` // [时间戳, 值]
}

// VMMatrixResult 矩阵结果
type VMMatrixResult []struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"` // [[时间戳, 值], ...]
}

// NewVMExecutor 创建 VictoriaMetrics 执行器
func NewVMExecutor(config config.VMConfig) (*VMExecutor, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("VictoriaMetrics URL is required")
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	return &VMExecutor{
		client:  client,
		baseURL: strings.TrimSuffix(config.URL, "/"),
		config:  config,
	}, nil
}

// Name 执行器名称
func (e *VMExecutor) Name() string {
	return "victoriaMetrics"
}

// Execute 执行巡检项
func (e *VMExecutor) Execute(ctx context.Context, item model.TaskItem) (model.TaskResult, error) {
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
		timeout = 30
	}
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 构建查询 URL
	queryURL := fmt.Sprintf("%s/prometheus/api/v1/query", e.baseURL)
	req, err := http.NewRequestWithContext(ctxWithTimeout, "GET", queryURL, nil)
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// 添加Authorization头
	req.Header.Set("Authorization", "Bearer "+e.config.Token)

	// 添加查询参数
	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	// 执行查询
	resp, err := e.client.Do(req)
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to execute query: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to read response: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// 解析响应
	var vmResp VMQueryResponse
	if err := json.Unmarshal(body, &vmResp); err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to parse response: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}
	hlog.Infof("vm query response: %s", string(body))
	// 检查响应状态
	if vmResp.Status != "success" {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Query failed with status: %s", vmResp.Status)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("query failed with status: %s", vmResp.Status)
	}

	// 解析结果
	parsedValue, allValues, err := e.parseVMResult(vmResp.Data.ResultType, vmResp.Data.Result)
	if err != nil {
		result.Status = model.ResultStatusFailed
		result.Message = fmt.Sprintf("Failed to parse result: %v", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// 设置结果值
	result.Value = parsedValue

	// 检查阈值
	if thresholdStr, ok := item.Params["threshold"].(string); ok {
		threshold, err := strconv.ParseFloat(thresholdStr, 64)
		if err == nil {
			// 检查所有节点值
			if vmResp.Data.ResultType == "vector" && len(allValues) > 0 {
				// 记录超过阈值的节点
				var exceedingNodes []string
				var highestValue float64
				var highestInstance string

				for instance, valueStr := range allValues {
					if valueFloat, err := strconv.ParseFloat(valueStr, 64); err == nil {
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
					result.Details += fmt.Sprintf("\n\n超过阈值的节点：\n%s", strings.Join(exceedingNodes, "\n"))
				}
			} else {
				// 对于非向量结果，保持原来的逻辑
				if resultValue, err := strconv.ParseFloat(parsedValue, 64); err == nil {
					if resultValue > threshold {
						result.Status = model.ResultStatusWarning
						result.Message = fmt.Sprintf("值 %.2f 超过阈值 %.2f", resultValue, threshold)
					}
				}
			}
		}
	}

	// 设置详细信息
	var details strings.Builder
	details.WriteString(fmt.Sprintf("Query: %s\n", query))
	details.WriteString(fmt.Sprintf("Result Type: %s\n", vmResp.Data.ResultType))
	details.WriteString(fmt.Sprintf("Result: %s\n", string(vmResp.Data.Result)))
	result.Details = details.String()
	result.Duration = time.Since(startTime).Milliseconds()

	return result, nil
}

// parseVMResult 解析 VictoriaMetrics 结果
func (e *VMExecutor) parseVMResult(resultType string, resultData json.RawMessage) (string, map[string]string, error) {
	allValues := make(map[string]string)

	switch resultType {
	case "scalar":
		var scalarResult VMScalarResult
		if err := json.Unmarshal(resultData, &scalarResult); err != nil {
			return "", allValues, err
		}
		if len(scalarResult) >= 2 {
			if value, ok := scalarResult[1].(string); ok {
				allValues["scalar"] = value
				return value, allValues, nil
			}
		}
		return string(resultData), allValues, nil

	case "vector":
		var vectorResult VMVectorResult
		if err := json.Unmarshal(resultData, &vectorResult); err != nil {
			return "", allValues, err
		}

		if len(vectorResult) == 0 {
			return "No data", allValues, nil
		}

		// 收集所有实例的值
		for _, item := range vectorResult {
			instance := "unknown"
			if instanceName, ok := item.Metric["instance"]; ok {
				instance = instanceName
			} else {
				// 如果没有instance标签，使用其他标签的组合作为标识
				parts := make([]string, 0, len(item.Metric))
				for k, v := range item.Metric {
					parts = append(parts, fmt.Sprintf("%s=%s", k, v))
				}
				if len(parts) > 0 {
					instance = strings.Join(parts, ",")
				}
			}

			if len(item.Value) >= 2 {
				if value, ok := item.Value[1].(string); ok {
					allValues[instance] = value
				}
			}
		}

		// 返回第一个值用于向后兼容
		if len(vectorResult) > 0 && len(vectorResult[0].Value) >= 2 {
			if value, ok := vectorResult[0].Value[1].(string); ok {
				return value, allValues, nil
			}
		}
		return string(resultData), allValues, nil

	case "matrix":
		var matrixResult VMMatrixResult
		if err := json.Unmarshal(resultData, &matrixResult); err != nil {
			return "", allValues, err
		}

		// 收集所有实例的最新值
		for _, item := range matrixResult {
			instance := "unknown"
			if instanceName, ok := item.Metric["instance"]; ok {
				instance = instanceName
			} else {
				// 如果没有instance标签，使用其他标签的组合作为标识
				parts := make([]string, 0, len(item.Metric))
				for k, v := range item.Metric {
					parts = append(parts, fmt.Sprintf("%s=%s", k, v))
				}
				if len(parts) > 0 {
					instance = strings.Join(parts, ",")
				}
			}

			if len(item.Values) > 0 && len(item.Values[len(item.Values)-1]) >= 2 {
				if value, ok := item.Values[len(item.Values)-1][1].(string); ok {
					allValues[instance] = value
				}
			}
		}

		if len(matrixResult) > 0 && len(matrixResult[0].Values) > 0 && len(matrixResult[0].Values[0]) >= 2 {
			if value, ok := matrixResult[0].Values[0][1].(string); ok {
				return value, allValues, nil
			}
		}
		return string(resultData), allValues, nil

	case "string":
		var stringResult string
		if err := json.Unmarshal(resultData, &stringResult); err != nil {
			return "", allValues, err
		}
		allValues["string"] = stringResult
		return stringResult, allValues, nil

	default:
		return string(resultData), allValues, nil
	}
}
