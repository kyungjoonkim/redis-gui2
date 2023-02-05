package rmodel

type RedisScanResult struct {
	Success      bool     `json:"success"`
	ErrorMessage string   `json:"errorMessage"`
	Keys         []string `json:"keys"`
	Cursor       int64    `json:"cursor"`
	Finish       bool     `json:"finish"`
}
