package configstorage

import (
	"example.com/sarang-apis/configmodel"
)

type ConfigStorageService interface {
	CreateConfig(data *configmodel.Config) (*configmodel.Config, error)
	UpdateConfig(configName string, configEnv configmodel.ConfigEnv, revisions []configmodel.ConfigRevision) error
	GetConfigByName(configName string, configEnv configmodel.ConfigEnv) (*configmodel.Config, error)
	GetAllConfig() ([]*configmodel.Config, error)
}
