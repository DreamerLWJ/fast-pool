package config

type DataSourceType int8

const (
	DataSourceTypeMySQL DataSourceType = iota + 1
	DataSourceTypeRedis
	DataSourceTypeZk
)
