package fabricsdk

import (
	"dendrix.io/fabricsdk/chaincode"
	"dendrix.io/fabricsdk/configs"
	"dendrix.io/fabricsdk/providers"
)

type fabricNetwork struct {
	cfgOptions     configs.ConfigOptions
	clientProvider providers.FabricNetworkClientProvider
}

//FabricNetwork defines the available fabric network methods
type FabricNetwork interface {
	ChaincodeInstallClient(clientOrgID string, chaincodeID string, chaincodeVersion string, chaincodePath string) (chaincode.ChaincodeClient, error)
	ChaincodeInstantiateClient(clientOrgID string, channelID string, chaincodeID string, chaincodeVersion string, chaincodePath string, policy string, args [][]byte, collectionConfigFile string) (chaincode.ChaincodeClient, error)
	ChaincodeUpgradeClient(clientOrgID string, channelID string, chaincodeID string, chaincodeVersion string, chaincodePath string, policy string, args [][]byte, collectionConfigFile string) (chaincode.ChaincodeClient, error)
	ChaincodeExecutionClient(clientOrgID string, channelID string, chaincodeID string, fn string, args [][]byte) (chaincode.ChaincodeClient, error)
	ChaincodeQueryClient(clientOrgID string, channelID string, chaincodeID string, fn string, args [][]byte) (chaincode.ChaincodeClient, error)
}

var fabNetwork *fabricNetwork

//NewFabricNetwork returns an instance of the fabric network
func NewFabricNetwork(configPath string) (FabricNetwork, error) {
	if err := initialize(configPath); err != nil {
		return nil, err
	}
	return fabNetwork, nil
}

func initialize(configPath string) error {
	//Get Network config options and store in memory
	//
	cfgOptions, err := configs.NewConfigOptions(configPath)
	if err != nil {
		return err
	}
	if fabNetwork == nil {
		fabNetwork = new(fabricNetwork)
	}
	fabNetwork.cfgOptions = cfgOptions
	return nil
}

func (fN *fabricNetwork) ChaincodeInstallClient(clientOrgID string, chaincodeID string, chaincodeVersion string, chaincodePath string) (chaincode.ChaincodeClient, error) {
	//Get the Client provider
	fNClientProvider := providers.NewFabricNetworkClientProvider(clientOrgID, fN.cfgOptions)
	//Get the chaincode client
	client, err := chaincode.NewInstallClient(fNClientProvider, chaincodeID, chaincodeVersion, chaincodePath)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (fN *fabricNetwork) ChaincodeInstantiateClient(clientOrgID string, channelID string, chaincodeID string, chaincodeVersion string, chaincodePath string, policy string, args [][]byte, collectionConfigFile string) (chaincode.ChaincodeClient, error) {
	//Get the Client provider
	fNClientProvider := providers.NewFabricNetworkClientProvider(clientOrgID, fN.cfgOptions)
	//Get the chaincode client
	client, err := chaincode.NewInstantiateClient(fNClientProvider, channelID, chaincodeID, chaincodeVersion, chaincodePath, policy, args, collectionConfigFile)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (fN *fabricNetwork) ChaincodeUpgradeClient(clientOrgID string, channelID string, chaincodeID string, chaincodeVersion string, chaincodePath string, policy string, args [][]byte, collectionConfigFile string) (chaincode.ChaincodeClient, error) {
	//Get the Client provider
	fNClientProvider := providers.NewFabricNetworkClientProvider(clientOrgID, fN.cfgOptions)
	//Get the chaincode client
	client, err := chaincode.NewUpgradeClient(fNClientProvider, channelID, chaincodeID, chaincodeVersion, chaincodePath, policy, args, collectionConfigFile)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (fN *fabricNetwork) ChaincodeExecutionClient(clientOrgID string, channelID string, chaincodeID string, fn string, args [][]byte) (chaincode.ChaincodeClient, error) {
	//Get the Client provider
	fNClientProvider := providers.NewFabricNetworkClientProvider(clientOrgID, fN.cfgOptions)
	//Get the chaincode client
	client := chaincode.NewExecuteClient(fNClientProvider, channelID, chaincodeID, fn, args)
	return client, nil
}

func (fN *fabricNetwork) ChaincodeQueryClient(clientOrgID string, channelID string, chaincodeID string, fn string, args [][]byte) (chaincode.ChaincodeClient, error) {
	//Get the Client provider
	fNClientProvider := providers.NewFabricNetworkClientProvider(clientOrgID, fN.cfgOptions)
	//Get the chaincode client
	client := chaincode.NewQueryClient(fNClientProvider, channelID, chaincodeID, fn, args)
	return client, nil
}
