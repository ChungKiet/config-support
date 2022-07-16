package configstorage

import (
	"context"
	"errors"

	"example.com/sarang-apis/configmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConfigEnv string
type ConfigState string

type ConfigCollection struct {
	configCollection *mongo.Collection
	ctx              context.Context
}

func NewUserService(cfgCollection *mongo.Collection, ctx context.Context) *ConfigCollection {
	return &ConfigCollection{
		configCollection: cfgCollection,
		ctx:              ctx,
	}
}

//CreateConfig create new config
func (cfgCol *ConfigCollection) CreateConfig(data *configmodel.Config) (*configmodel.Config, error) {
	_, err := cfgCol.configCollection.InsertOne(cfgCol.ctx, data)

	return data, err
}

// UpdateConfig update one config by name and env
func (cfgCol *ConfigCollection) UpdateConfig(configName string, configEnv configmodel.ConfigEnv, revisions []configmodel.ConfigRevision) error {

	filter := bson.M{"name": configName, "env": configEnv}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "name", Value: configName}, primitive.E{Key: "env", Value: configEnv}, primitive.E{Key: "revisions", Value: revisions}}}}
	_, err := cfgCol.configCollection.UpdateOne(cfgCol.ctx, filter, update)
	return err
}

// GetConfigByName view one config by name, env and id revision
func (cfgCol *ConfigCollection) GetConfigByName(configName string, configEnv configmodel.ConfigEnv) (*configmodel.Config, error) {
	data := &configmodel.Config{}

	// var user *models.User
	query := bson.D{bson.E{Key: "name", Value: configName}, bson.E{Key: "env", Value: configEnv}}
	err := cfgCol.configCollection.FindOne(cfgCol.ctx, query).Decode(&data)

	// Find config by name and env

	if data.Name == "" {
		return data, nil
	}

	if err != nil {
		return nil, errors.New("Cannot get config")
	}

	return data, nil
}

// GetAllConfig get all configs
func (cfgCol *ConfigCollection) GetAllConfig() ([]*configmodel.Config, error) {
	var allConfigs []*configmodel.Config

	// Get all configs
	cursor, err := cfgCol.configCollection.Find(cfgCol.ctx, bson.D{{}})
	for cursor.Next(cfgCol.ctx) {
		var config configmodel.Config
		err := cursor.Decode(&config)
		if err != nil {
			return nil, err
		}
		// Append config in map
		allConfigs = append(allConfigs, &config)
	}

	err = cursor.Close(cfgCol.ctx)

	if err != nil {
		return nil, err
	}

	// If there are no config, then return nil
	if len(allConfigs) == 0 {
		return nil, errors.New("no documents found")
	}

	return allConfigs, nil
}
