package configmodel

import (
	"time"
)

type ConfigEnv string
type ConfigState string

const (
	EnvDev     ConfigEnv = "DEV"
	EnvStaging ConfigEnv = "STAGING"
	EnvUAT     ConfigEnv = "UAT"
	EnvProd    ConfigEnv = "PRODUCTION"

	StateApproved   ConfigState = "approved"
	StateUnapproved ConfigState = "unapproved"

	AddRev            string = "yes"
	ErrEditPermission int    = 5102
	CreatePermission  int    = 4900
	Zero              int    = 0
	InputInvalid      int    = 5000
)

type (
	// ConfigRevision
	//		ID			: ID Revision
	//		Value		: value of revision (exp: "{"test": 1}")
	//		CreatedAt	: Time create revision
	//		UpdatedAt	: Time update revision
	//		Author		: The revision's author
	//		State		: 'approved' or 'unapproved'
	ConfigRevision struct {
		ID        int         `json:"id" bson:"id"`
		Value     string      `json:"value" bson:"value"`
		CreatedAt time.Time   `json:"createdAt" bson:"createdAt"`
		UpdatedAt time.Time   `json:"updatedAt" bson:"updatedAt"`
		Author    string      `json:"author" bson:"author"`
		State     ConfigState `json:"state" bson:"state"`
	}

	//Config
	//		mongodb.DefaultModel	: id of config (mongo id default)
	//		ConfigName				: name of config
	//	 	Env						: environment of config ('DEV' / 'STAGING' / 'UAT' / 'PRODUCTION' )
	// 		Revisions				: include many revisions
	Config struct {
		Name      string           `json:"name" bson:"name"`
		Env       ConfigEnv        `json:"env" bson:"env"`
		Revisions []ConfigRevision `json:"revisions" bson:"revisions"`
	}
)
