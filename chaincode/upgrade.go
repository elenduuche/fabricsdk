package chaincode

import (
	"fmt"
	"strings"

	"dendrix.io/fabricsdk/providers"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

type upgradeChaincodeClient struct {
	//To indicate that this interface is implemented
	ChaincodeClient
	providers.FabricNetworkClientProvider
	channelID        string
	chaincodeID      string
	chaincodePath    string
	chaincodeVersion string
	policy           string
	args             [][]byte
	collConfig       []*common.CollectionConfig
}

//NewUpgradeClient returns a ChaincodeClient implementation for upgrading a chaincode on the client org anchor peer
func NewUpgradeClient(provider providers.FabricNetworkClientProvider, channelID string, chaincodeID string, chaincodeVersion string, chaincodePath string, policy string, args [][]byte, collectionConfigFile string) (ChaincodeClient, error) {
	i := new(upgradeChaincodeClient)
	i.FabricNetworkClientProvider = provider
	i.channelID = channelID
	i.chaincodeID = chaincodeID
	i.chaincodeVersion = chaincodeVersion
	i.chaincodePath = chaincodePath
	i.policy = policy
	i.args = args
	// Private Data Collection Configuration
	// - see fixtures/config/pvtdatacollection.json for sample config file
	collCfg, err := collectionConfig(collectionConfigFile)
	if err != nil {
		return nil, err
	}
	i.collConfig = collCfg
	return i, nil
}

func (ic upgradeChaincodeClient) Invoke() ([]byte, error) {
	resMgmtClient, err := ic.ResourceMgmtClientByAdmin()
	if err != nil {
		return []byte("0x00"), err
	}
	ccIDVersion := ic.chaincodeID + "." + ic.chaincodeVersion
	fmt.Printf("\n Sending upgrade %s ...\n", ccIDVersion)

	chaincodePolicy, err := ic.newChaincodePolicy()
	if err != nil {
		return []byte("0x00"), err
	}

	req := resmgmt.UpgradeCCRequest{
		Name:       ic.chaincodeID,
		Path:       ic.chaincodePath,
		Version:    ic.chaincodeVersion,
		Args:       ic.args,
		Policy:     chaincodePolicy,
		CollConfig: ic.collConfig,
	}

	_, err = resMgmtClient.UpgradeCC(ic.channelID, req, resmgmt.WithTargets(ic.ClientOrgPeers()[0]))
	if err != nil {
		if strings.Contains(err.Error(), "chaincode exists "+ic.chaincodeID) {
			// Ignore
			//cliconfig.Config().Logger().Infof("Chaincode %s already instantiated.", cliconfig.Config().ChaincodeID())
			fmt.Printf("...chaincode %s already instantiated.\n", ic.chaincodeID)
			return []byte("EXISTS"), nil
		}
		return nil, errors.Errorf("error upgrading chaincode: %v", err)
	}

	fmt.Printf("...successfuly upgraded chaincode %s on channel %s.\n", ic.chaincodeID, ic.channelID)
	return []byte("OK"), nil
}

func (ic upgradeChaincodeClient) Terminate() {
	ic.CloseSDK()
}

func (ic upgradeChaincodeClient) newChaincodePolicy() (*common.SignaturePolicyEnvelope, error) {
	if ic.policy != "" {
		// Create a signature policy from the policy expression passed in
		return newChaincodePolicy(ic.policy)
	}

	// Default policy is 'signed by any member' for all known orgs
	return cauthdsl.AcceptAllPolicy, nil
}
