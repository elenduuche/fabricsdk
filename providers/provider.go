package providers

import (
	"fmt"

	"dendrix.io/fabricsdk/configs"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

type chaincodeClient struct {
	clientProvider
}

type channelClient struct {
}

//FabricNetworkClientProvider define methods implemented by the fabric network client provider
type FabricNetworkClientProvider interface {
	ResourceMgmtClient() (*resmgmt.Client, error)
	ResourceMgmtClientByAdmin() (*resmgmt.Client, error)
	ResourceMgmtClientByUser(username string) (*resmgmt.Client, error)
	ResourceMgmtClientByOrg(username string, orgID string) (*resmgmt.Client, error)
	ClientOrgID() string
	ParticipatingOrgs() []string
	ClientOrgMSPID() string
	ClientAdminUser() mspapi.SigningIdentity
	ClientUser() mspapi.SigningIdentity
	ClientOrgPeers() []fab.Peer
	PeersByOrgID() map[string][]fab.Peer
	CloseSDK()
	ChannelClient(channelID string) (*channel.Client, error)
}

//clientProvider provides the fabric network context for a client organisation
type clientProvider struct {
	cfgOptions      configs.ConfigOptions
	sdk             *fabsdk.FabricSDK
	peers           []fab.Peer
	orgIDByPeer     map[string]string
	orgsID          []string
	orgMSPID        string
	sessions        map[string]context.ClientProvider
	channelSessions map[string]context.ChannelProvider
	userName        string
	user            mspapi.SigningIdentity
	adminUser       mspapi.SigningIdentity
	clientOrgID     string
	peersByOrg      map[string][]fab.Peer
}

//NewFabricNetworkClientProvider return an instance of the client Org's Fabric Network ClientProvider
func NewFabricNetworkClientProvider(clientOrgID string, cfgOptions configs.ConfigOptions) FabricNetworkClientProvider {
	clientProvider := new(clientProvider)
	//clientProvider.cfgOptions = cfgOptions
	clientProvider.sdk = cfgOptions.GetFabricSDK(clientOrgID)
	clientProvider.peers = cfgOptions.GetClientOrgPeers(clientOrgID)
	clientProvider.orgIDByPeer = cfgOptions.GetOrgsIDByPeers(clientOrgID)
	clientProvider.orgsID = cfgOptions.GetOrgsID(clientOrgID)
	clientProvider.userName = cfgOptions.GetUserName(clientOrgID)
	clientProvider.user = cfgOptions.GetClientOrgUser(clientOrgID)
	clientProvider.adminUser = cfgOptions.GetClientOrgAdminUser(clientOrgID)
	clientProvider.orgMSPID = cfgOptions.GetClientOrgMSPID(clientOrgID)
	clientProvider.peersByOrg = cfgOptions.GetAllPeersByOrg(clientOrgID)
	return clientProvider
}

//ResourceMgmtClient returns the resmgmt.Client for the org user
func (cProv *clientProvider) ResourceMgmtClient() (*resmgmt.Client, error) {
	//Get resmgmt client
	session, err := cProv.context(cProv.user)
	if err != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve context clientprovider: %s", err.Error())
	}
	resmgmtClient, err1 := resmgmt.New(session)
	if err1 != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve resmgmt client for: %s", err.Error())
	}
	return resmgmtClient, nil
}

//ResourceMgmtClientByAdmin returns the resmgmt.Client for the org admin
func (cProv *clientProvider) ResourceMgmtClientByAdmin() (*resmgmt.Client, error) {
	//Get resmgmt client
	session, err := cProv.context(cProv.adminUser)
	if err != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve context clientprovider: %s", err.Error())
	}
	resmgmtClient, err1 := resmgmt.New(session)
	if err1 != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve resmgmt client for: %s", err.Error())
	}
	return resmgmtClient, nil
}

//ResourceMgmtClientByUser returns the resmgmt.Client for a specified org username
func (cProv *clientProvider) ResourceMgmtClientByUser(username string) (*resmgmt.Client, error) {
	//Get resmgmt client
	user, err0 := cProv.mspUser(username)
	if err0 != nil {
		return nil, err0
	}
	session, err := cProv.context(user)
	if err != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve context clientprovider: %s", err.Error())
	}
	resmgmtClient, err1 := resmgmt.New(session)
	if err1 != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve resmgmt client for: %s", err.Error())
	}
	return resmgmtClient, nil
}

//ResourceMgmtClientByOrg returns the resmgmt.Client for a specified org and username
func (cProv *clientProvider) ResourceMgmtClientByOrg(username string, orgID string) (*resmgmt.Client, error) {
	//Get resmgmt client
	user, err0 := cProv.mspUserByOrg(username, orgID)
	if err0 != nil {
		return nil, err0
	}
	session, err := cProv.context(user)
	if err != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve context client provider for user %s of org %s Error: %s", user.Identifier().ID, user.Identifier().MSPID, err.Error())
	}
	resmgmtClient, err1 := resmgmt.New(session)
	if err1 != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve resmgmt client for: %s", err.Error())
	}
	return resmgmtClient, nil
}

func (cProv *clientProvider) context(user mspapi.SigningIdentity) (context.ClientProvider, error) {
	key := user.Identifier().MSPID + "_" + user.Identifier().ID
	session := cProv.sessions[key]
	if session == nil {
		session = cProv.sdk.Context(fabsdk.WithIdentity(user))
		cProv.sessions[key] = session
	}
	return session, nil
}

func (cProv *clientProvider) channelContext(user mspapi.SigningIdentity, channelID string) (context.ChannelProvider, error) {
	key := user.Identifier().MSPID + "_" + user.Identifier().ID + "_" + channelID
	session := cProv.channelSessions[key]
	if session == nil {
		session = cProv.sdk.ChannelContext(channelID, fabsdk.WithIdentity(user))
		cProv.channelSessions[key] = session
	}
	return session, nil
}

func (cProv *clientProvider) ClientOrgID() string {
	return cProv.clientOrgID
}

func (cProv *clientProvider) ParticipatingOrgs() []string {
	return cProv.orgsID
}

func (cProv *clientProvider) ClientOrgMSPID() string {
	return cProv.orgMSPID
}

func (cProv *clientProvider) ClientAdminUser() mspapi.SigningIdentity {
	return cProv.adminUser
}

func (cProv *clientProvider) ClientUser() mspapi.SigningIdentity {
	return cProv.user
}

func (cProv *clientProvider) ClientUserName() string {
	return cProv.userName
}

func (cProv *clientProvider) ClientOrgPeers() []fab.Peer {
	return cProv.peers
}

func (cProv *clientProvider) PeersByOrgID() map[string][]fab.Peer {
	return cProv.peersByOrg
}

func (cProv *clientProvider) CloseSDK() {
	if cProv.sdk != nil {
		fmt.Println("Closing SDK")
		cProv.sdk.Close()
	}
}

//ChannelClient returns the channel.Client for the org user
func (cProv *clientProvider) ChannelClient(channelID string) (*channel.Client, error) {
	//Get resmgmt client
	session, err := cProv.channelContext(cProv.user, channelID)
	if err != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve context channel provider for channel: %s. Error - %s", channelID, err.Error())
	}
	channelClient, err1 := channel.New(session)
	if err1 != nil {
		return nil, errors.Errorf("Error occurred when attempting to retrieve channel client for channel: %s. Error - %s", channelID, err.Error())
	}
	return channelClient, nil
}

func (cProv *clientProvider) mspUser(username string) (mspapi.SigningIdentity, error) {
	mspClient, err := msp.New(cProv.sdk.Context(), msp.WithOrg(cProv.clientOrgID))
	if err != nil {
		return nil, errors.Errorf("error creating MSP client for %s. Error: %v", username, err)
	}
	user, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		return nil, errors.Errorf("GetSigningIdentity for %s returned error: %v", username, err)
	}
	return user, nil
}

func (cProv *clientProvider) mspUserByOrg(username string, orgID string) (mspapi.SigningIdentity, error) {
	mspClient, err := msp.New(cProv.sdk.Context(), msp.WithOrg(orgID))
	if err != nil {
		return nil, errors.Errorf("error creating MSP client for %s of org %s. Error: %v", username, orgID, err)
	}
	user, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		return nil, errors.Errorf("GetSigningIdentity for %s of org %s returned error: %v", username, orgID, err)
	}
	return user, nil
}
