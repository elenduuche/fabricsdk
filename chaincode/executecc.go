package chaincode

import (
	"math/rand"

	"dendrix.io/fabricsdk/providers"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

//Invoke invokes chaincode
/* func (setup *FabricSetup) Invoke(function string, invokeArgs [][]byte) (string, error) {
	//fmt.Println("@Invoke -> Chaincode:", setup.ChainCodeID, "Function:", function, "Args length:", len(invokeArgs))
	setup.CreateChannelClient()

	peerIndex := 0
	numberOfpeers := len(setup.Peers)
	if numberOfpeers > 1 {
		peerIndex = rand.Intn(numberOfpeers)
	}

	response, err := setup.client.Execute(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: function, Args: invokeArgs}, channel.WithTargets(setup.Peers[peerIndex]))

	if err != nil {
		return "", fmt.Errorf("Failed to invoke function: %v", err)
	}

	return string(response.Payload), nil
} */

type executeChaincodeClient struct {
	//To indicate that this interface is implemented
	ChaincodeClient
	providers.FabricNetworkClientProvider
	channelID   string
	chaincodeID string
	args        [][]byte
	function    string
}

//NewExecuteClient returns a ChaincodeClient implmentation for executing chaincode business functions
func NewExecuteClient(provider providers.FabricNetworkClientProvider, channelID string, chaincodeID string, fn string, args [][]byte) ChaincodeClient {
	i := new(executeChaincodeClient)
	i.FabricNetworkClientProvider = provider
	i.channelID = channelID
	i.chaincodeID = chaincodeID
	i.args = args
	i.function = fn
	return i
}

func (ic executeChaincodeClient) Invoke() ([]byte, error) {
	chClient, err := ic.ChannelClient(ic.channelID)
	if err != nil {
		return []byte("0x00"), err
	}

	req := channel.Request{
		ChaincodeID: ic.chaincodeID,
		Fcn:         ic.function,
		Args:        ic.args,
	}
	target := ic.peer(ic.ClientOrgPeers())
	response, err := chClient.Execute(req, channel.WithTargets(target))

	if err != nil {
		return nil, errors.Errorf("failed to invoke function %s on chaincode %s. Error - %s", req.Fcn, req.ChaincodeID, err.Error())
	}

	return response.Payload, nil
}

func (ic executeChaincodeClient) Terminate() {
	ic.CloseSDK()
}

func (ic executeChaincodeClient) peer(targets []fab.Peer) fab.Peer {
	peerIndex := 0
	numberOfpeers := len(targets)
	if numberOfpeers > 1 {
		peerIndex = rand.Intn(numberOfpeers)
	}
	return targets[peerIndex]
}
