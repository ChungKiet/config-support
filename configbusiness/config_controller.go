package configbusiness

import (
	"context"
	"encoding/json"
	"example.com/sarang-apis/config-pubsub/publisher"
	"fmt"
	"net/http"

	"time"

	"example.com/sarang-apis/configmodel"
	"example.com/sarang-apis/configstorage"
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type configPubSub struct {
	ConfigName string                `json:"configName"`
	Env        configmodel.ConfigEnv `json:"env"`
	IdRevision int                   `json:"idRevision"`
}

const (
	maxWorker  = 64
	CONFIG_PUB = "mongo-config"
)

// configAdapter
//		Name		: name of config
//		Env			: environment of config ('DEV' / 'STAGING' / 'UAT' / 'PRODUCTION')
//		IDRevision	: id of revision
//		Value		: value of this revision (exp: "{"test": 1"})
//		Author		: author of this revision
//		State		: 'approved' or 'unapproved'
//		Copy		: 'copy' or 'no'
type configAdapter struct {
	Name       string                  `form:"name"`
	Env        configmodel.ConfigEnv   `form:"env"`
	IDRevision int                     `form:"idRevision"`
	Value      string                  `form:"value"`
	Author     string                  `form:"author"`
	State      configmodel.ConfigState `form:"state"`
	Add        string                  `form:"add"`
}

// getConfig
//		Name	: name of config
//		Env		: environment of config
//		Rev		: id of revision
type getConfig struct {
	Name       string                `form:"name"`
	Env        configmodel.ConfigEnv `form:"env"`
	IDRevision int                   `form:"idRevision"`
}

type ConfigController struct {
	ConfigService configstorage.ConfigStorageService
	workerPool    *workerpool.WorkerPool
}

func New(configService configstorage.ConfigStorageService) ConfigController {
	return ConfigController{
		ConfigService: configService,
		workerPool:    workerpool.New(maxWorker),
	}
}

// CreateConfig create new config:
// configEnv: dev / staging / uat / production
// value: JSON.Marshal(Config) ("{"test": 1")
// author: config creator
func (c *ConfigController) CreateConfig(ctx *gin.Context) {

	config := configmodel.Config{}
	request := &configAdapter{}
	if err := ctx.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Check whether the name exists
	cfg, err := c.ConfigService.GetConfigByName(request.Name, request.Env)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "Can't create this config1"})
		return
	}

	if cfg.Name != "" {
		ctx.JSON(http.StatusOK, gin.H{"message": "Can't create this config"})
		return
	}

	// Init new config
	config.Name = request.Name
	config.Env = request.Env
	revision := configmodel.ConfigRevision{}

	// Append an empty revision to config.Revisions
	config.Revisions = append(config.Revisions, revision)

	// ID Revision start at 1
	config.Revisions[configmodel.Zero].ID = 1
	config.Revisions[configmodel.Zero].Author = request.Author
	config.Revisions[configmodel.Zero].CreatedAt = time.Now()

	// The state init is 'unapproved'
	config.Revisions[configmodel.Zero].State = configmodel.StateUnapproved
	config.Revisions[configmodel.Zero].Value = request.Value

	// Create new config
	result, err := c.ConfigService.CreateConfig(&config)

	if err != nil || result == nil {
		fmt.Print("Error create config")
		ctx.JSON(http.StatusOK, err)
	}

	ctx.JSON(http.StatusOK, result)
}

// UpdateConfig
// Update config value if param is valid
func (c *ConfigController) UpdateConfig(ctx *gin.Context) {

	// lock by config name

	request := configAdapter{}
	if err := ctx.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	//lock, err := redislock.LockCustomRetry(request.Name, 5)
	//if err != nil {
	//	fmt.Print(err.Error())
	//	return
	//}
	//
	//defer func() {
	//	_, errUnlock := redislock.Unlock(lock)
	//	if errUnlock != nil {
	//		log.Println("checkStatus unlock err", errUnlock.Error())
	//	}
	//}()

	// check whether the config name is exists
	cfg, err := c.ConfigService.GetConfigByName(request.Name, request.Env)

	// return null if the name was not found
	if err != nil || cfg.Name == "" {
		ctx.JSON(http.StatusBadRequest, "Config does not exists")
		return
	}

	// If rev = 0, set rev = len(Revisions). It's the latest revision
	if request.IDRevision == 0 && request.Add != configmodel.AddRev {
		request.IDRevision = getLatestIDByState(cfg, configmodel.StateUnapproved)
	} else if request.IDRevision == 0 {
		request.IDRevision = getMaxIDRev(cfg) + 1
	}

	if request.IDRevision == 0 && request.Add != configmodel.AddRev {
		ctx.JSON(http.StatusBadRequest, "All revision was approved")
		return
	}

	// check if rev is not nil

	cfgRev := getRevByID(cfg, request.IDRevision)
	if len(cfgRev) == 0 && request.Add != configmodel.AddRev {
		ctx.JSON(http.StatusBadRequest, "Bad request (id revision is invalid)")
		return
	}

	// Check whether this config is approved and copy = "no"
	// If copy = "yes" then create new config and change value
	if request.Add != configmodel.AddRev && cfgRev[0].State == configmodel.StateApproved {
		ctx.JSON(http.StatusBadRequest, "You can't edit this config")
		return
	}

	revision := configmodel.ConfigRevision{}
	// If copy = 'yes' and id is valid, set revision = config revision where id = i
	if request.Add == configmodel.AddRev && request.IDRevision != getMaxIDRev(cfg)+1 {
		revision = cfgRev[0]
	}

	if request.Add == configmodel.AddRev {
		// Change state to unapproved
		revision.State = configmodel.StateUnapproved
		revision.CreatedAt = time.Now()
		// Increase ID of new revision
		maxRevID := getMaxIDRev(cfg)
		revision.ID = maxRevID + 1
		// Append new revision
		cfg.Revisions = append(cfg.Revisions, revision)
		request.IDRevision = maxRevID + 1
	}

	checkAndUpdate(cfg, request.IDRevision, request.State, request.Author, request.Value)

	// Update config by name and env
	err = c.ConfigService.UpdateConfig(cfg.Name, cfg.Env, cfg.Revisions)

	if request.State == configmodel.StateApproved {
		c.PublishMongoConfigMessage(request.Name, request.Env, request.IDRevision)
	}

	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
	}

	ctx.JSON(http.StatusOK, cfg)
}

// GetConfigByName : Get config by name, env and id revision
func (c *ConfigController) GetConfigByName(ctx *gin.Context) {

	// Check whether config is exists
	request := &getConfig{}

	if err := ctx.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	cfg, err := c.ConfigService.GetConfigByName(request.Name, request.Env)
	if err != nil || cfg.Name == "" {
		ctx.JSON(http.StatusBadRequest, "Config does not exists")
		return
	}

	// If rev = -1, then return all of revisions
	if request.IDRevision == -1 {
		ctx.JSON(http.StatusOK, cfg)
	}

	// Check whether revision's id is valid
	// If rev = 0, set rev = len(Revisions). It's the latest revision
	if request.IDRevision == 0 {
		request.IDRevision = getLatestIDByState(cfg, configmodel.StateApproved)
	}

	// check if rev is not nil
	cfgRev := getRevByID(cfg, request.IDRevision)
	if len(cfgRev) == 0 {
		ctx.JSON(http.StatusBadRequest, "Bad request (id revision is invalid)")
		return
	}
	// return config with exactly one revision
	result := cfg
	result.Revisions = cfgRev

	ctx.JSON(http.StatusOK, cfg)
}

// GetAllConfig : Return all of configs
func (c *ConfigController) GetAllConfig(ctx *gin.Context) {

	// Get all of configs
	totalConfig, err := c.ConfigService.GetAllConfig()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, totalConfig)
}

// getLatestIDByState : Get the latest ID of revision by state (approved or unapproved)
func getLatestIDByState(m *configmodel.Config, state configmodel.ConfigState) int {
	id := 0
	for _, rev := range m.Revisions {
		if rev.State == state && rev.ID > id {
			id = rev.ID
		}
	}

	// return 0, if there is no config match with state
	return id
}

// getRevByID: get revision of config by id
func getRevByID(cfg *configmodel.Config, rev int) []configmodel.ConfigRevision {
	var result []configmodel.ConfigRevision
	for _, cfgRev := range cfg.Revisions {
		if cfgRev.ID == rev {
			result = append(result, cfgRev)
			break
		}
	}
	// return config with one revision
	return result
}

// setRevByID: set revision of config by id
func setRevByID(cfg *configmodel.Config, rev int, cfgRev configmodel.ConfigRevision) {
	for i, cRev := range cfg.Revisions {
		if cRev.ID == rev {
			// find revision with id = i and set to new revision
			cfg.Revisions[i] = cfgRev
			return
		}
	}
}

// getMaxIDRev: get max id in revisions
func getMaxIDRev(cfg *configmodel.Config) int {
	var result int = 0
	for _, cfgRev := range cfg.Revisions {
		if cfgRev.ID > result {
			result = cfgRev.ID
		}
	}
	return result
}

// checkAndUpdate: check if value is not nil, update config
func checkAndUpdate(config *configmodel.Config, rev int, state configmodel.ConfigState, author string, value string) {
	cfgRev := getRevByID(config, rev)

	// if state is valid, update state
	if state == configmodel.StateApproved || state == configmodel.StateUnapproved {
		cfgRev[0].State = state
	}

	// if author is valid, update author
	if author != "" {
		cfgRev[0].Author = author
	}

	// if value is valid, update value
	if value != "" {
		cfgRev[0].Value = value
	}

	cfgRev[0].UpdatedAt = time.Now()
	// set revision of config by id
	setRevByID(config, rev, cfgRev[0])
}

func (c *ConfigController) RegisterUserRoutes(rg *gin.RouterGroup) {
	configroute := rg.Group("/config")
	configroute.GET("/view-config", c.GetConfigByName)
	configroute.GET("/view-all-config", c.GetAllConfig)
	configroute.PUT("/update-config", c.UpdateConfig)
	configroute.POST("/create-config", c.CreateConfig)

}

// pushConfigApproved: publish message when new config is approved
func pushConfigApproved(ctx context.Context, configName string, configEnv configmodel.ConfigEnv, idRevision int) error {
	// Message include: config name, env and id of revision
	message := configPubSub{
		ConfigName: configName,
		Env:        configEnv,
		IdRevision: idRevision,
	}

	// marshal message
	raw, err := json.Marshal(message)

	if err != nil {
		return err
	}

	fmt.Print(raw)

	// publish message
	err = publisher.PublishMessage(ctx, CONFIG_PUB, raw)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigController) PublishMongoConfigMessage(configName string, configEnv configmodel.ConfigEnv, idRevision int) error {

	c.workerPool.Submit(func() {
		err := pushConfigApproved(context.Background(), configName, configEnv, idRevision)
		if err != nil {
			// TODO Log
			fmt.Print(err.Error(), "-PublishMongoConfigMessage-", "configName-", configName)
		}
	})
	return nil
}
