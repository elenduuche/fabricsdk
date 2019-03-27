package configs

import (
	fabapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

type configOptionService struct {
	appCfgMap     map[string]*appConfig
	networkCfgMap map[string]*networkConfig
}

//ConfigOptions struct defines the config properties of the app/chaincode and fabric network
type ConfigOptions interface {
	GetUserName(clientOrgID string) string
	GetFabricSDK(clientOrgID string) *fabsdk.FabricSDK
	GetClientOrgID(clientOrgID string) string
	GetClientOrgMSPID(clientOrgID string) string
	GetClientOrgPeers(clientOrgID string) []fabapi.Peer
	GetClientOrgUser(clientOrgID string) mspapi.SigningIdentity
	GetClientOrgAdminUser(clientOrgID string) mspapi.SigningIdentity
	GetOrgsID(clientOrgID string) []string
	GetOrgsMSPByOrgID(clientOrgID string) map[string]string
	GetOrgsIDByPeers(clientOrgID string) map[string]string
	GetClientOrgs() []string
	GetAllPeersByOrg(clientOrgID string) map[string][]fabapi.Peer
}

//NewConfigOptions initializes the ConfigOptions struct
func NewConfigOptions(configPath string) (ConfigOptions, error) {
	cfgOptions := new(configOptionService)
	//Init App config
	err := cfgOptions.initAppCfg(configPath)
	if err != nil {
		return nil, errors.Errorf("Initialization of App Config failed with the error: %s", err.Error())
	}
	//Init organisations
	//Init peers
	//Init channel
	//Init identities
	err = cfgOptions.initNetworkCfg()
	if err != nil {
		return nil, errors.Errorf("Initialization of Network Config failed with the error: %s", err.Error())
	}
	return cfgOptions, nil
}

func (copts *configOptionService) GetUserName(clientOrgID string) string {
	return copts.appCfgMap[clientOrgID].getUser()
}

/* func (copts *configOptionService) GetPolicy() (*common.SignaturePolicyEnvelope, error) {
	if copts.appCfg.getPolicy() != "" {
		return newChaincodePolicy(copts.appCfg.getPolicy())
	}
	return cauthdsl.AcceptAllPolicy, nil
} */

func (copts *configOptionService) GetFabricSDK(clientOrgID string) *fabsdk.FabricSDK {
	return copts.networkCfgMap[clientOrgID].sdk
}

func (copts *configOptionService) GetClientOrgID(clientOrgID string) string {
	return copts.networkCfgMap[clientOrgID].clientOrgID
}

func (copts *configOptionService) GetClientOrgMSPID(clientOrgID string) string {
	return copts.networkCfgMap[clientOrgID].clientOrgMSPID
}

func (copts *configOptionService) GetClientOrgPeers(clientOrgID string) []fabapi.Peer {
	return copts.networkCfgMap[clientOrgID].clientOrgPeers
}

func (copts *configOptionService) GetClientOrgUser(clientOrgID string) mspapi.SigningIdentity {
	return copts.networkCfgMap[clientOrgID].clientOrgUser
}

func (copts *configOptionService) GetClientOrgAdminUser(clientOrgID string) mspapi.SigningIdentity {
	return copts.networkCfgMap[clientOrgID].clientOrgAdminUser
}

func (copts *configOptionService) GetOrgsID(clientOrgID string) []string {
	return copts.networkCfgMap[clientOrgID].orgsID
}

func (copts *configOptionService) GetOrgsMSPByOrgID(clientOrgID string) map[string]string {
	return copts.networkCfgMap[clientOrgID].orgsMSPByOrgID
}

func (copts *configOptionService) GetOrgsIDByPeers(clientOrgID string) map[string]string {
	return copts.networkCfgMap[clientOrgID].orgsIDByPeers
}

func (copts *configOptionService) GetClientOrgs() []string {
	var orgs []string
	for org := range copts.appCfgMap {
		orgs = append(orgs, org)
	}
	return orgs
}

func (copts *configOptionService) GetAllPeersByOrg(clientOrgID string) map[string][]fabapi.Peer {
	return copts.networkCfgMap[clientOrgID].peersByOrg
}

func (copts *configOptionService) initAppCfg(appConfigPath string) error {
	appConfigMap, err := initAppConfig(appConfigPath)
	if err != nil {
		return err
	}
	copts.appCfgMap = appConfigMap
	return nil
}

func (copts *configOptionService) initNetworkCfg() error {
	if copts.appCfgMap == nil {
		return errors.New("App Config was not initialized")
	}
	copts.networkCfgMap = make(map[string]*networkConfig)
	for orgid, appCfg := range copts.appCfgMap {
		networkConfig, err := initNetworkConfig(appCfg.getNetworkConfigPath(), appCfg.getUser())
		if err != nil {
			return err
		}
		copts.networkCfgMap[orgid] = networkConfig
	}
	return nil
}

func newChaincodePolicy(policyString string) (*common.SignaturePolicyEnvelope, error) {
	ccPolicy, err := cauthdsl.FromString(policyString)
	if err != nil {
		return nil, errors.Errorf("invalid chaincode policy [%s]: %s", policyString, err)
	}
	return ccPolicy, nil
}
