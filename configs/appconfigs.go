package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

//List of App config keys
const (
	// chaincodePath     = "chaincodepath"
	// chaincodeVersion  = "version"
	// chaincodeID       = "chaincodeid"
	username          = "user"
	networkConfigPath = "networkconfigpath"
	// policyType        = "policytype"
	// policySigners     = "policysigners"
	// policySignerRole  = "policysignerrole"
	// policy            = "policy"
	clientOrgs = "clientorgs"
	orgID      = "orgid"
)

/* const (
	POLICYTYPE_OR  = "ANY"
	POLICYTYPE_AND = "SPECIFIC"
) */

const appConfigFile = "fabricApp.json"

type appConfig map[string]interface{}

/* func (cfg appConfig) getChaincodePath() string {
	return cfg[chaincodePath].(string)
}

func (cfg appConfig) getChaincodeVersion() string {
	return cfg[chaincodeVersion].(string)
}

func (cfg appConfig) getChaincodeID() string {
	return cfg[chaincodeID].(string)
} */

func (cfg appConfig) getUser() string {
	return cfg[username].(string)
}

func (cfg appConfig) getNetworkConfigPath() string {
	return cfg[networkConfigPath].(string)
}

func (cfg appConfig) getOrgID() string {
	return cfg[orgID].(string)
}

/* func (cfg appConfig) getPolicyType() string {
	return cfg[policyType].(string)
}

func (cfg appConfig) getPolicySigners() []string {
	return cfg[policySigners].([]string)
}

func (cfg appConfig) getPolicySignerRole() string {
	return cfg[policySignerRole].(string)
}

func (cfg appConfig) getPolicy() string {
	return cfg[policy].(string)
} */

//getAppConfig creates and initializes an instance of appConfig by loading the fabricApp.json config
func initAppConfig(appConfigPath string) (map[string]*appConfig, error) {
	v := viper.New()
	// v.SetEnvPrefix(envPrefix)
	// v.BindEnv(configEnvVar)
	v.AddConfigPath(appConfigPath)
	v.SetConfigFile(appConfigFile)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var configMap = make(map[string]*appConfig)
	clientorgs := v.Get(clientOrgs).([]interface{})
	for _, o := range clientorgs {
		org := o.(string)
		fmt.Println(org)
		val := v.Get(org)
		if val != nil {
			c := appConfig(val.(map[string]interface{}))
			configMap[org] = &c
		}
	}
	return configMap, nil
}
