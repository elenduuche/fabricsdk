package chaincode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"dendrix.io/fabricsdk/providers"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

type instantiateChaincodeClient struct {
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

//NewInstantiateClient returns a ChaincodeClient implmentation for instantiating a chaincode on the client org anchor peer
func NewInstantiateClient(provider providers.FabricNetworkClientProvider, channelID string, chaincodeID string, chaincodeVersion string, chaincodePath string, policy string, args [][]byte, collectionConfigFile string) (ChaincodeClient, error) {
	if provider == nil {
		return nil, errors.Errorf("Fabric network client provider is not set.")
	}
	i := new(instantiateChaincodeClient)
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

func (ic instantiateChaincodeClient) Invoke() ([]byte, error) {
	resMgmtClient, err := ic.ResourceMgmtClientByAdmin()
	if err != nil {
		return []byte("0x00"), err
	}
	ccIDVersion := ic.chaincodeID + "." + ic.chaincodeVersion
	fmt.Printf("\n Sending instantiate %s ...\n", ccIDVersion)

	chaincodePolicy, err := ic.newChaincodePolicy()
	if err != nil {
		return []byte("0x00"), err
	}

	req := resmgmt.InstantiateCCRequest{
		Name:       ic.chaincodeID,
		Path:       ic.chaincodePath,
		Version:    ic.chaincodeVersion,
		Args:       ic.args,
		Policy:     chaincodePolicy,
		CollConfig: ic.collConfig,
	}

	_, err = resMgmtClient.InstantiateCC(ic.channelID, req, resmgmt.WithTargets(ic.ClientOrgPeers()[0]))
	if err != nil {
		if strings.Contains(err.Error(), "chaincode exists "+ic.chaincodeID) {
			// Ignore
			//cliconfig.Config().Logger().Infof("Chaincode %s already instantiated.", cliconfig.Config().ChaincodeID())
			fmt.Printf("...chaincode %s already instantiated.\n", ic.chaincodeID)
			return []byte("EXISTS"), nil
		}
		return nil, errors.Errorf("error instantiating chaincode: %v", err)
	}

	fmt.Printf("...successfuly instantiated chaincode %s on channel %s.\n", ic.chaincodeID, ic.channelID)
	return []byte("OK"), nil
}

func (ic instantiateChaincodeClient) Terminate() {
	ic.CloseSDK()
}

func (ic instantiateChaincodeClient) newChaincodePolicy() (*common.SignaturePolicyEnvelope, error) {
	if ic.policy != "" {
		// Create a signature policy from the policy expression passed in
		return newChaincodePolicy(ic.policy)
	}

	// Default policy is 'signed by any member' for all known orgs
	return cauthdsl.AcceptAllPolicy, nil
}

func newChaincodePolicy(policyString string) (*common.SignaturePolicyEnvelope, error) {
	ccPolicy, err := cauthdsl.FromString(policyString)
	if err != nil {
		return nil, errors.Errorf("invalid chaincode policy [%s]: %s", policyString, err)
	}
	return ccPolicy, nil
}

func getCollectionConfigFromFile(ccFile string) ([]*common.CollectionConfig, error) {
	fileBytes, err := ioutil.ReadFile(ccFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read file [%s]", ccFile)
	}
	cconf := &[]collectionConfigJSON{}
	if err = json.Unmarshal(fileBytes, cconf); err != nil {
		return nil, errors.Wrapf(err, "error parsing collection configuration in file [%s]", ccFile)
	}
	return getCollectionConfig(*cconf)
}

func getCollectionConfig(cconf []collectionConfigJSON) ([]*common.CollectionConfig, error) {
	ccarray := make([]*common.CollectionConfig, 0, len(cconf))
	for _, cconfitem := range cconf {
		p, err := cauthdsl.FromString(cconfitem.Policy)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("invalid policy %s", cconfitem.Policy))
		}
		cpc := &common.CollectionPolicyConfig{
			Payload: &common.CollectionPolicyConfig_SignaturePolicy{
				SignaturePolicy: p,
			},
		}
		cc := &common.CollectionConfig{
			Payload: &common.CollectionConfig_StaticCollectionConfig{
				StaticCollectionConfig: &common.StaticCollectionConfig{
					Name:              cconfitem.Name,
					MemberOrgsPolicy:  cpc,
					RequiredPeerCount: cconfitem.RequiredPeerCount,
					MaximumPeerCount:  cconfitem.MaxPeerCount,
				},
			},
		}
		ccarray = append(ccarray, cc)
	}
	return ccarray, nil
}

func collectionConfig(filePath string) ([]*common.CollectionConfig, error) {
	var collConfig []*common.CollectionConfig
	var err error
	collConfigFile := filePath
	if collConfigFile != "" {
		collConfig, err = getCollectionConfigFromFile(filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting private data collection configuration from file [%s]", filePath)
		}
		return collConfig, nil
	}
	return nil, nil
}

type collectionConfigJSON struct {
	Name              string `json:"name"`
	Policy            string `json:"policy"`
	RequiredPeerCount int32  `json:"requiredPeerCount"`
	MaxPeerCount      int32  `json:"maxPeerCount"`
}
