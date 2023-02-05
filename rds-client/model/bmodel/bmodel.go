package bmodel

type RedisGetParamModel struct {
	RedisKey string
	Start    int64
	End      int64
}

func NewRedisGetParamModel(redisKey string, start int64, end int64) *RedisGetParamModel {
	result := &RedisGetParamModel{}
	result.RedisKey = redisKey

	if start > 0 {
		result.Start = start
	}

	if end <= 0 {
		result.End = 1000
	} else {
		result.End = end
	}

	return result
}

type RedisGetResModel struct {
	DataType string      `json:"dataType"`
	RedisKey string      `json:"redisKey"`
	Next     int64       `json:"next"`
	Values   interface{} `json:"values"`
}
