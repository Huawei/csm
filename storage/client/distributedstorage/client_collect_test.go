/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package distributedstorage

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	apiDistributed "github.com/huawei/csm/v2/storage/api/distributedstorage"
	"github.com/huawei/csm/v2/storage/client"
)

// --- groupByObjectId tests ---

func TestGroupByObjectId_SingleObject(t *testing.T) {
	// arrange
	data := []interface{}{
		map[string]interface{}{
			"id":               "35",
			"indicator":        "30003",
			"indicator_values": []interface{}{"0", "1"},
		},
		map[string]interface{}{
			"id":               "35",
			"indicator":        "30002",
			"indicator_values": []interface{}{"0"},
		},
	}
	client := &DistributedClient{}

	// act
	result := client.groupByObjectId(data)

	// assert
	assert.Len(t, result, 1)
	assert.Equal(t, "35", result[0]["object_id"])
	assert.Equal(t, []int{30003, 30002}, result[0]["indicators"])
	assert.Equal(t, []float64{1, 0}, result[0]["indicator_values"]) // takes latest value (last element)
}

func TestGroupByObjectId_MultipleObjects(t *testing.T) {
	// arrange
	data := []interface{}{
		map[string]interface{}{
			"id":               "35",
			"indicator":        "30001",
			"indicator_values": []interface{}{"1024"},
		},
		map[string]interface{}{
			"id":               "36",
			"indicator":        "30001",
			"indicator_values": []interface{}{"2048"},
		},
	}
	client := &DistributedClient{}

	// act
	result := client.groupByObjectId(data)

	// assert
	assert.Len(t, result, 2)
	assert.Equal(t, "35", result[0]["object_id"])
	assert.Equal(t, []float64{1024}, result[0]["indicator_values"])
	assert.Equal(t, "36", result[1]["object_id"])
	assert.Equal(t, []float64{2048}, result[1]["indicator_values"])
}

func TestGroupByObjectId_EmptyData(t *testing.T) {
	// arrange
	data := []interface{}{}
	client := &DistributedClient{}

	// act
	result := client.groupByObjectId(data)

	// assert
	assert.Empty(t, result)
}

func TestGroupByObjectId_MissingIndicatorValues(t *testing.T) {
	// arrange
	data := []interface{}{
		map[string]interface{}{
			"id":        "35",
			"indicator": "30001",
			// no indicator_values
		},
	}
	client := &DistributedClient{}

	// act
	result := client.groupByObjectId(data)

	// assert
	assert.Empty(t, result)
}

func TestGroupByObjectId_MissingID(t *testing.T) {
	// arrange
	data := []interface{}{
		map[string]interface{}{
			"indicator":        "30001",
			"indicator_values": []interface{}{"1024"},
			// no id
		},
	}
	client := &DistributedClient{}

	// act
	result := client.groupByObjectId(data)

	// assert
	assert.Empty(t, result)
}

func TestGroupByObjectId_EmptyIndicatorValues(t *testing.T) {
	// arrange
	data := []interface{}{
		map[string]interface{}{
			"id":               "35",
			"indicator":        "30001",
			"indicator_values": []interface{}{},
		},
	}
	client := &DistributedClient{}

	// act
	result := client.groupByObjectId(data)

	// assert
	assert.Empty(t, result)
}

// --- GetPageByUrlKey range parameter encoding test ---

func TestGetPageByUrlKey_RangeParameter(t *testing.T) {
	// arrange — verify the range JSON encoding matches expected format
	rangeParam := map[string]int{"offset": 0, "limit": 100}
	rangeJSON, err := json.Marshal(rangeParam)

	// act & assert
	assert.NoError(t, err)
	assert.Equal(t, `{"limit":100,"offset":0}`, string(rangeJSON))
}

// --- GetSingleByUrlKey tests ---

func TestDistributedClient_GetSingleByUrlKey_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects/detail", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   map[string]interface{}{"id": float64(1), "name": "obj1"},
			}, nil
		})
	defer patches.Reset()

	// action
	gotData, gotErr := dc.GetSingleByUrlKey(context.Background(), "GetController")

	// assert
	assert.Nil(t, gotErr)
	assert.NotNil(t, gotData)
	assert.Equal(t, float64(1), gotData["ID"])
	assert.Equal(t, "obj1", gotData["NAME"])
}

func TestDistributedClient_GetSingleByUrlKey_GenerateUrlFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}
	wantErr := errors.New("generate url failed")

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "", wantErr)
	defer patches.Reset()

	// action
	gotData, gotErr := dc.GetSingleByUrlKey(context.Background(), "GetController")

	// assert
	assert.Nil(t, gotData)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "generate url failed")
}

func TestDistributedClient_GetSingleByUrlKey_CallFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}
	wantErr := errors.New("connection refused")

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects/detail", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return nil, wantErr
		})
	defer patches.Reset()

	// action
	gotData, gotErr := dc.GetSingleByUrlKey(context.Background(), "GetController")

	// assert
	assert.Nil(t, gotData)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "get single by url key")
}

func TestDistributedClient_GetSingleByUrlKey_DataFieldNotMap(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects/detail", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   []interface{}{"not-a-map"},
			}, nil
		})
	defer patches.Reset()

	// action
	gotData, gotErr := dc.GetSingleByUrlKey(context.Background(), "GetController")

	// assert
	assert.Nil(t, gotData)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "data field can not convert to map")
}

// --- GetListByUrlKey tests ---

func TestDistributedClient_GetListByUrlKey_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data": []interface{}{
					map[string]interface{}{"id": float64(1), "name": "obj1"},
					map[string]interface{}{"id": float64(2), "name": "obj2"},
				},
			}, nil
		})
	defer patches.Reset()

	// action
	gotList, gotErr := dc.GetListByUrlKey(context.Background(), "GetControllers")

	// assert
	assert.Nil(t, gotErr)
	assert.Len(t, gotList, 2)
	assert.Equal(t, float64(1), gotList[0]["ID"])
	assert.Equal(t, "obj2", gotList[1]["NAME"])
}

func TestDistributedClient_GetListByUrlKey_DataFieldNotArray(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   "not-an-array",
			}, nil
		})
	defer patches.Reset()

	// action
	gotList, gotErr := dc.GetListByUrlKey(context.Background(), "GetControllers")

	// assert
	assert.Nil(t, gotList)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "data field can not convert to array")
}

// --- GetCountByUrlKey tests ---

func TestDistributedClient_GetCountByUrlKey_Success_float64Count(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects/count", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   map[string]interface{}{"count": float64(42)},
			}, nil
		})
	defer patches.Reset()

	// action
	gotCount, gotErr := dc.GetCountByUrlKey(context.Background(), "GetLunCount")

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, 42, gotCount)
}

func TestDistributedClient_GetCountByUrlKey_CountFieldNotFound(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects/count", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   map[string]interface{}{"total": float64(10)},
			}, nil
		})
	defer patches.Reset()

	// action
	gotCount, gotErr := dc.GetCountByUrlKey(context.Background(), "GetLunCount")

	// assert
	assert.Equal(t, 0, gotCount)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "count field can not convert to int")
}

func TestDistributedClient_GetCountByUrlKey_InvalidStringCount(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects/count", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   map[string]interface{}{"count": "not-a-number"},
			}, nil
		})
	defer patches.Reset()

	// action
	gotCount, gotErr := dc.GetCountByUrlKey(context.Background(), "GetLunCount")

	// assert
	assert.Equal(t, 0, gotCount)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "count field can not convert to int")
}

// --- GetPageByUrlKey tests ---

func TestDistributedClient_GetPageByUrlKey_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects/page", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data": []interface{}{
					map[string]interface{}{"id": float64(1), "name": "obj1"},
					map[string]interface{}{"id": float64(2), "name": "obj2"},
				},
			}, nil
		})
	defer patches.Reset()

	// action
	gotPage, gotErr := dc.GetPageByUrlKey(context.Background(), "GetLuns", 0, 100)

	// assert
	assert.Nil(t, gotErr)
	assert.Len(t, gotPage, 2)
	assert.Equal(t, float64(1), gotPage[0]["ID"])
}

func TestDistributedClient_GetPageByUrlKey_CallFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}
	wantErr := errors.New("connection refused")

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/objects/page", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return nil, wantErr
		})
	defer patches.Reset()

	// action
	gotPage, gotErr := dc.GetPageByUrlKey(context.Background(), "GetLuns", 0, 100)

	// assert
	assert.Nil(t, gotPage)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "get page by url key")
}

// --- QueryPerformanceData tests ---

func TestDistributedClient_QueryPerformanceData_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
		ESN: "test-esn",
	}
	// Register object type for the test
	RegisterObjectType(100, "GetControllers", "", "", "ID")

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/test", nil)

	callCount := 0
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			callCount++
			if callCount == 1 {
				// getSystemTime call
				return map[string]interface{}{
					"result": map[string]interface{}{"code": float64(0)},
					"data":   map[string]interface{}{"CMO_SYS_UTC_TIME": float64(1700000000)},
				}, nil
			}
			if callCount == 2 {
				// GetListByUrlKey call (getObjectIdsFromList)
				return map[string]interface{}{
					"result": map[string]interface{}{"code": float64(0)},
					"data": []interface{}{
						map[string]interface{}{"ID": float64(1)},
						map[string]interface{}{"ID": float64(2)},
					},
				}, nil
			}
			// POST performance_data call
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data": []interface{}{
					map[string]interface{}{
						"id":               "1",
						"indicator":        "30001",
						"indicator_values": []interface{}{"1024"},
					},
				},
			}, nil
		})
	defer patches.Reset()

	// action
	gotResult, gotErr := dc.QueryPerformanceData(context.Background(), 100, []string{"30001"})

	// assert
	assert.Nil(t, gotErr)
	assert.Len(t, gotResult, 1)
	assert.Equal(t, "1", gotResult[0]["object_id"])
}

func TestDistributedClient_QueryPerformanceData_ObjectTypeNotRegistered(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/system/time", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   map[string]interface{}{"CMO_SYS_UTC_TIME": float64(1700000000)},
			}, nil
		})
	defer patches.Reset()

	// action — use an unregistered object type
	gotResult, gotErr := dc.QueryPerformanceData(context.Background(), 99999, []string{"30001"})

	// assert
	assert.Nil(t, gotResult)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "is not registered in objectTypeMapping")
}

func TestDistributedClient_QueryPerformanceData_AllIndicatorsInvalid(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
		ESN: "test-esn",
	}
	RegisterObjectType(101, "GetControllers", "", "", "ID")

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/test", nil)

	callCount := 0
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			callCount++
			if callCount == 1 {
				// getSystemTime call
				return map[string]interface{}{
					"result": map[string]interface{}{"code": float64(0)},
					"data":   map[string]interface{}{"CMO_SYS_UTC_TIME": float64(1700000000)},
				}, nil
			}
			// GetListByUrlKey call (getObjectIdsFromList)
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data": []interface{}{
					map[string]interface{}{"ID": float64(1)},
				},
			}, nil
		})
	defer patches.Reset()

	// action — pass all-invalid indicators
	gotResult, gotErr := dc.QueryPerformanceData(context.Background(), 101, []string{"invalid", "also-invalid"})

	// assert
	assert.Nil(t, gotResult)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "no valid indicators after conversion")
}

func TestDistributedClient_QueryPerformanceData_GetSystemTimeFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
	}
	wantErr := errors.New("network error")

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/system/time", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string, _ string,
			_ map[string]interface{}) (map[string]interface{}, error) {
			return nil, wantErr
		})
	defer patches.Reset()

	// action
	gotResult, gotErr := dc.QueryPerformanceData(context.Background(), 100, []string{"30001"})

	// assert
	assert.Nil(t, gotResult)
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "get system time failed")
}
