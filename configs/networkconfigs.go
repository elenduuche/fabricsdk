package configs

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	fabapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const adminUser = "Admin"

type networkConfig struct {
	sdk                *fabsdk.FabricSDK
	endpointCfg        fabapi.EndpointConfig
	identityCfg        mspapi.IdentityConfig
	clientOrgID        string
	clientOrgMSPID     string
	clientOrgPeers     []fabapi.Peer
	clientOrgUser      mspapi.SigningIdentity
	clientOrgAdminUser mspapi.SigningIdentity
	orgsID             []string
	orgsMSPByOrgID     map[string]string
	orgsIDByPeers      map[string]string
	peersByOrg         map[string][]fabapi.Peer
}

func getNetworkConfig(networkConfigPath string) (*networkConfig, error) {
	var netCfg = new(networkConfig)
	networkCfgProvider := config.FromFile(networkConfigPath)
	sdk, err := fabsdk.New(networkCfgProvider)
	if err != nil {
		return nil, err
	}
	clientProvider := sdk.Context()
	ctx, err1 := clientProvider()
	if err1 != nil {
		return nil, err1
	}
	endpointCfg := ctx.EndpointConfig()
	identityCfg := ctx.IdentityConfig()
	netCfg.sdk = sdk
	netCfg.endpointCfg = endpointCfg
	netCfg.identityCfg = identityCfg
	return nil, nil
}

func initNetworkConfig(networkConfigPath string, username string) (*networkConfig, error) {
	netCfg, err := getNetworkConfig(networkConfigPath)
	if err != nil {
		return nil, errors.Errorf("Network config initialization failed with error: %s", err.Error())
	}
	netCfg.initClientOrg()
	if err := netCfg.initClientOrgMSPID(); err != nil {
		return nil, errors.Errorf("Network config initialization failed with error: %s", err.Error())
	}
	if err := netCfg.initClientOrgPeers(); err != nil {
		return nil, errors.Errorf("Network config initialization failed with error: %s", err.Error())
	}
	if err := netCfg.initClientOrgUser(username); err != nil {
		return nil, errors.Errorf("Network config initialization failed with error: %s", err.Error())
	}
	if err := netCfg.initClientOrgAdminUser(); err != nil {
		return nil, errors.Errorf("Network config initialization failed with error: %s", err.Error())
	}
	if err := netCfg.initOrgs(); err != nil {
		return nil, errors.Errorf("Network config initialization failed with error: %s", err.Error())
	}
	if err := netCfg.initOrgsIDByPeers(); err != nil {
		return nil, errors.Errorf("Network config initialization failed with error: %s", err.Error())
	}
	if err := netCfg.initParticipatingOrgPeers(); err != nil {
		return nil, errors.Errorf("Network config initialization failed with error: %s", err.Error())
	}
	return netCfg, nil
}

func (netCfg *networkConfig) initClientOrg() {
	netCfg.clientOrgID = netCfg.identityCfg.Client().Organization
}

func (netCfg *networkConfig) initClientOrgMSPID() error {
	orgCfgMap := netCfg.endpointCfg.NetworkConfig().Organizations
	orgCfg := orgCfgMap[netCfg.clientOrgID]
	if &orgCfg == nil {
		errMsg := fmt.Sprintf("OrgID: %s is invalid. Failed to find the MSPID.", netCfg.clientOrgID)
		return errors.New(errMsg)
	}
	netCfg.clientOrgMSPID = orgCfg.MSPID
	return nil
}

func (netCfg *networkConfig) initClientOrgPeers() error {
	peerCfg, found := netCfg.endpointCfg.PeersConfig(netCfg.clientOrgID)
	if !found {
		errMsg := fmt.Sprintf("OrgID: %s is invalid. Failed to find the PeerConfigs.", netCfg.clientOrgID)
		return errors.New(errMsg)
	}
	ctx, err := netCfg.sdk.Context()()
	if err != nil {
		return nil
	}
	for _, p := range peerCfg {
		peer, err := ctx.InfraProvider().CreatePeerFromConfig(&fabapi.NetworkPeer{PeerConfig: p})
		if err != nil {
			return err
		}
		netCfg.clientOrgPeers = append(netCfg.clientOrgPeers, peer)
	}
	return nil
}

func (netCfg *networkConfig) initClientOrgUser(username string) error {
	mspClient, err := msp.New(netCfg.sdk.Context(), msp.WithOrg(netCfg.clientOrgID))
	if err != nil {
		return errors.Errorf("error creating MSP client: %s", err)
	}
	user, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		return errors.Errorf("GetSigningIdentity for %s returned error: %v", username, err)
	}
	netCfg.clientOrgUser = user
	return nil
}

func (netCfg *networkConfig) initClientOrgAdminUser() error {
	username := adminUser
	mspClient, err := msp.New(netCfg.sdk.Context(), msp.WithOrg(netCfg.clientOrgID))
	if err != nil {
		return errors.Errorf("error creating MSP client: %s", err)
	}
	admin, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		return errors.Errorf("GetSigningIdentity for %s returned error: %v", username, err)
	}
	netCfg.clientOrgAdminUser = admin
	return nil
}

func (netCfg *networkConfig) initOrgs() error {
	orgCfgMap := netCfg.endpointCfg.NetworkConfig().Organizations
	if netCfg.orgsMSPByOrgID == nil {
		netCfg.orgsMSPByOrgID = make(map[string]string)
	}
	for orgID, orgCfg := range orgCfgMap {
		if !strings.Contains(orgID, "orderer") {
			netCfg.orgsID = append(netCfg.orgsID, orgID)
			netCfg.orgsMSPByOrgID[orgID] = orgCfg.MSPID
		}
	}
	return nil
}

func (netCfg *networkConfig) initOrgsIDByPeers() error {
	orgCfgMap := netCfg.endpointCfg.NetworkConfig().Organizations
	if netCfg.orgsIDByPeers == nil {
		netCfg.orgsIDByPeers = make(map[string]string)
	}
	for orgID, orgCfg := range orgCfgMap {
		if !strings.Contains(orgID, "orderer") {
			netCfg.orgsID = append(netCfg.orgsID, orgID)
			for _, p := range orgCfg.Peers {
				netCfg.orgsIDByPeers[p] = orgID
			}
		}
	}
	return nil
}

func (netCfg *networkConfig) initParticipatingOrgPeers() error {

	if len(netCfg.orgsID) == 0 {
		return errors.Errorf("@initParticipatingOrgPeers - network Config Options 'Org ID' slice is not set")
	}

	var peersByOrg = make(map[string][]fabapi.Peer)

	for _, orgid := range netCfg.orgsID {
		peerCfg, found := netCfg.endpointCfg.PeersConfig(orgid)
		if !found {
			return errors.Errorf("@initParticipatingOrgPeers - OrgID: %s is invalid. Failed to find the PeerConfigs.", orgid)

		}
		ctx, err := netCfg.sdk.Context()()
		if err != nil {
			return nil
		}
		for _, p := range peerCfg {
			peer, err := ctx.InfraProvider().CreatePeerFromConfig(&fabapi.NetworkPeer{PeerConfig: p})
			if err != nil {
				return err
			}
			peersByOrg[orgid] = append(peersByOrg[orgid], peer)
		}
	}
	netCfg.peersByOrg = peersByOrg
	return nil
}
