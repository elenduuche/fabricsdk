package chaincode

import (
	"dendrix.io/fabricsdk/providers"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

//Query queries chaincode
/* func (setup *FabricSetup) Query(function string, args [][]byte) (string, error) {
	setup.CreateChannelClient()
	response, err := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: function, Args: args})
	if err != nil {
		return "", fmt.Errorf("Failed to query: %v", err)
	}

	return string(response.Payload), nil
} */

type queryChaincodeClient struct {
	//To indicate that this interface is implemented
	ChaincodeClient
	providers.FabricNetworkClientProvider
	channelID   string
	chaincodeID string
	args        [][]byte
	function    string
}

//NewQueryClient returns a ChaincodeClient implmentation for querying chaincode business functions
func NewQueryClient(provider providers.FabricNetworkClientProvider, channelID string, chaincodeID string, fn string, args [][]byte) ChaincodeClient {
	i := new(queryChaincodeClient)
	i.FabricNetworkClientProvider = provider
	i.channelID = channelID
	i.chaincodeID = chaincodeID
	i.args = args
	i.function = fn
	return i
}

func (ic queryChaincodeClient) Invoke() ([]byte, error) {
	chClient, err := ic.ChannelClient(ic.channelID)
	if err != nil {
		return []byte("0x00"), err
	}

	req := channel.Request{
		ChaincodeID: ic.chaincodeID,
		Fcn:         ic.function,
		Args:        ic.args,
	}
	//target := ic.peer(ic.ClientOrgPeers())
	response, err := chClient.Execute(req)

	if err != nil {
		return nil, errors.Errorf("failed to query function %s on chaincode %s. Error - %s", req.Fcn, req.ChaincodeID, err.Error())
	}
	return response.Payload, nil
}

func (ic queryChaincodeClient) Terminate() {
	ic.CloseSDK()
}
