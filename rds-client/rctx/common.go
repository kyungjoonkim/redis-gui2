package rctx

import (
	"changeme/rds-client/client"
	"changeme/rds-client/model/bmodel"
	result_model "changeme/rds-client/model/rmodel"
	"fmt"
	"golang.org/x/net/context"
	"strconv"
)

const (
	String = "string"
	List   = "list"
	Set    = "set"
	Zset   = "zset"
	Hash   = "hash"
	Stream = "stream"
)

var TypeMap = map[string]func(*client.Client, *bmodel.RedisGetParamModel) *bmodel.RedisGetResModel{
	String: getRedisStringTypeData,
	Zset:   getRedisZRANGETypeData,
	Hash:   getRedisHSCANTypeData,
}

func LoginRedisServer(address string, password string) (RedisClientContext, error) {

	basicClient, err := client.Connect(context.Background(), address, password)
	if err != nil {
		return nil, err
	}

	if isCluster(basicClient) {
		return createClusterClient(basicClient, password)
	}

	cmdCtx, err := createCommandContext(basicClient)
	if err != nil {
		return nil, err
	}

	return &SingleRedisClientContext{
		client:              basicClient,
		db:                  0,
		ipAndPort:           address,
		RedisCommandContext: cmdCtx,
	}, nil

}

func ScanRedisKeyInTargetNode(targetClient *client.Client, cursor int64) *result_model.RedisScanResult {
	scanResult, err := targetClient.Send([]interface{}{
		"scan",
		strconv.FormatInt(cursor, 10),
		"count",
		"800",
		"match",
		"*",
	})

	result := &result_model.RedisScanResult{}
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}

	scanSlice, ok := scanResult.Content().([]interface{})
	if !ok {
		result.ErrorMessage = err.Error()
		return result
	}

	for index, scanData := range scanSlice {
		if index == 0 {
			fmt.Println(scanData)
			parseCursor, err := strconv.ParseInt(scanData.(string), 10, 64)
			if err != nil {
				result.ErrorMessage = err.Error()
				return result
			}
			result.Cursor = parseCursor
			if result.Cursor == 0 {
				result.Finish = true
			}
			continue
		}
		keys, ok := scanData.([]interface{})
		if !ok {
			break
		}
		resultKeys := make([]string, 0)
		for _, key := range keys {
			resultKeys = append(resultKeys, key.(string))
		}

		result.Keys = resultKeys
	}

	result.Success = true

	return result
}

func slotNodes(slotNum int64, slots []*RedisClusterSlot) *RedisClusterSlot {
	for _, slot := range slots {
		if slotNum >= slot.start && slotNum <= slot.end {
			return slot
		}
	}
	return nil
}

func createCommandContext(client *client.Client) (*RedisCommandContext, error) {

	res, err := client.Send([]interface{}{
		"command",
	})

	if err != nil {
		return nil, err
	}

	resCmds, ok := res.Content().([]interface{})
	if !ok {
		return nil, err
	}

	commandMap := make(map[string]*RedisCommand)
	for _, resTargetCmds := range resCmds {
		resTargetCmdInfos, ok := resTargetCmds.([]interface{})
		if !ok {
			continue
		}
		resultCommand := &RedisCommand{}
		for index, resTargetCmdInfo := range resTargetCmdInfos {
			if index == 0 {
				resultCommand.name = resTargetCmdInfo.(string)
				continue
			}

			if index == 1 {
				resultCommand.arity = resTargetCmdInfo.(int64)
				continue
			}

			if index == 2 {
				resFlags, ok := resTargetCmdInfo.([]interface{})
				if ok {
					resultCommand.flags = make([]string, 0)
					for _, resFlag := range resFlags {
						resultCommand.flags = append(resultCommand.flags, resFlag.(string))
					}
				}
				continue
			}

			if index == 3 {
				resultCommand.firstKey = resTargetCmdInfo.(int64)
				continue
			}

			if index == 4 {
				resultCommand.lastKey = resTargetCmdInfo.(int64)
				continue
			}

			if index == 5 {
				resultCommand.step = resTargetCmdInfo.(int64)
				continue
			}

		}
		if len(resultCommand.name) > 0 {
			commandMap[resultCommand.name] = resultCommand
		}

	}

	result := &RedisCommandContext{
		redisCommandMap: commandMap,
	}

	return result, nil
}

func getSlotInfo(targetSlotInfoSlice []interface{}) *RedisClusterSlot {
	result := &RedisClusterSlot{}
	result.nodeInfos = make([]*ClusterNodeInfo, 0)

	for index, targetSlotInfo := range targetSlotInfoSlice {
		if index == 0 { //start
			result.start = targetSlotInfo.(int64)
		} else if index == 1 { //end
			result.end = targetSlotInfo.(int64)
		} else if index > 1 { //
			if nodeInfoSlice, ok := targetSlotInfo.([]interface{}); ok {
				resultNodeInfo := &ClusterNodeInfo{}
				for nodeIndex, nodeInfo := range nodeInfoSlice {
					if nodeIndex == 0 {
						resultNodeInfo.ip = nodeInfo.(string)
					} else if nodeIndex == 1 {
						resultNodeInfo.port = nodeInfo.(int64)
					} else if nodeIndex == 2 {
						resultNodeInfo.nodeID = nodeInfo.(string)

					}
				}
				result.nodeInfos = append(result.nodeInfos, resultNodeInfo)
			}

		}
	}
	return result
}

func createClusterClient(basicClient *client.Client, password string) (*ClusterRedisClientContext, error) {
	response, err := basicClient.Send([]interface{}{
		"cluster",
		"slots",
	})

	if err != nil {
		return nil, err
	}

	if _, errOk := response.Content().(error); errOk {
		return nil, err
	}

	slots, ok := response.Content().([]interface{})
	if !ok {
		return nil, err
	}

	redisClusterSlots := make([]*RedisClusterSlot, len(slots))
	nodeInfoMap := make(map[string]*ClusterNodeInfo)
	for index, targetSlotInfos := range slots {
		targetSlotInfoSlice, ok := targetSlotInfos.([]interface{})
		if !ok {
			continue
		}

		slotInfo := getSlotInfo(targetSlotInfoSlice)
		for _, nodeInfo := range slotInfo.nodeInfos {

			client, err := client.Connect(context.Background(), nodeInfo.ip+":"+strconv.FormatInt(nodeInfo.port, 10), password)
			if err != nil {
				continue
			}

			nodeInfo.client = client
			targetIpAndPort := nodeInfo.ip + ":" + strconv.FormatInt(nodeInfo.port, 10)
			nodeInfoMap[targetIpAndPort] = nodeInfo

		}
		redisClusterSlots[index] = slotInfo
	}

	cmdCtx, err := createCommandContext(basicClient)
	if err != nil {
		return nil, err
	}

	result := &ClusterRedisClientContext{
		ClusterNodeInfos:    redisClusterSlots,
		RedisCommandContext: cmdCtx,
		nodeInfoMap:         nodeInfoMap,
	}

	err = basicClient.Close()
	return result, nil
}

func isCluster(basicClient *client.Client) bool {
	_, err := basicClient.Send([]interface{}{
		"CLUSTER",
		"INFO",
	})

	if err == nil {
		return true
	}
	return false
}

func getRedisStringTypeData(client *client.Client, param *bmodel.RedisGetParamModel) *bmodel.RedisGetResModel {
	res, err := client.Send([]interface{}{
		"get",
		param.RedisKey,
	})

	if err != nil {
		return nil
	}

	return &bmodel.RedisGetResModel{
		DataType: String,
		RedisKey: param.RedisKey,
		Values:   res.Content(),
	}
}

func getRedisZRANGETypeData(client *client.Client, param *bmodel.RedisGetParamModel) *bmodel.RedisGetResModel {
	res, err := client.Send([]interface{}{
		"zrange",
		param.RedisKey,
		strconv.FormatInt(param.Start, 10),
		strconv.FormatInt(param.End, 10),
		"withscores",
	})

	if err != nil {
		return nil
	}
	resultList := res.ConvertList()
	length := len(resultList)
	restLength := (param.End + 1) - int64((length / 2))
	next := int64(-1)
	if restLength == 0 {
		next = param.End + 1
	}
	valueList := make([]interface{}, 0)
	for idx, _ := range resultList {
		if idx%2 == 0 {
			valueMap := make(map[string]interface{})
			valueMap["value"] = resultList[idx]
			valueMap["score"] = resultList[idx+1]
			valueList = append(valueList, valueMap)
		}

	}

	return &bmodel.RedisGetResModel{
		DataType: Zset,
		RedisKey: param.RedisKey,
		Values:   valueList,
		Next:     next,
	}
}

func getRedisHSCANTypeData(client *client.Client, param *bmodel.RedisGetParamModel) *bmodel.RedisGetResModel {
	res, err := client.Send([]interface{}{
		"hscan",
		param.RedisKey,
		strconv.FormatInt(param.Start, 10),
		"count",
		strconv.FormatInt(param.End, 10),
	})

	if err != nil {
		return nil
	}
	fmt.Println(res.Content())

	resultList := res.ConvertList()
	strNext := resultList[0].(string)
	next, _ := strconv.ParseInt(strNext, 10, 64)

	return &bmodel.RedisGetResModel{
		DataType: Hash,
		RedisKey: param.RedisKey,
		Values:   resultList[1],
		Next:     next,
	}
}

func getRedisKeyData(client *client.Client, param *bmodel.RedisGetParamModel) *bmodel.RedisGetResModel {

	result, err := client.Send([]interface{}{
		"type",
		param.RedisKey,
	})

	if err != nil {
		return nil
	}

	typeData := result.Content().(string)

	function := TypeMap[typeData]
	return function(client, param)

}
