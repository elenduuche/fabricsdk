package chaincode

import (
	"fmt"
	"net/http"
	"os"

	"dendrix.io/fabricsdk/providers"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/pkg/errors"
)

const admin = "Admin"

type installChaincodeClient struct {
	//To indicate that this interface is implemented
	ChaincodeClient
	providers.FabricNetworkClientProvider
	chaincodeID      string
	chaincodePath    string
	chaincodeVersion string
}

//NewInstallClient returns a ChaincodeClient implmentation for installing chaincode on the network peers
func NewInstallClient(provider providers.FabricNetworkClientProvider, chaincodeID string, chaincodeVersion string, chaincodePath string) (ChaincodeClient, error) {
	//Verify that the provider is not nil
	if provider == nil {
		return nil, errors.Errorf("Fabric network client provider is not set.")
	}
	i := new(installChaincodeClient)
	i.FabricNetworkClientProvider = provider
	i.chaincodeID = chaincodeID
	i.chaincodeVersion = chaincodeVersion
	i.chaincodePath = chaincodePath
	return i, nil
}

func (ic *installChaincodeClient) installChaincode(orgID string, targets []fab.Peer) error {
	//Package chaincode
	goPath := os.Getenv("GOPATH")
	ccPkg, err := gopackager.NewCCPackage(ic.chaincodePath, goPath)
	if err != nil {
		return err
	}
	//Create InstallCCRequest instance
	req := resmgmt.InstallCCRequest{
		Name:    ic.chaincodeID,
		Path:    ic.chaincodePath,
		Version: ic.chaincodeVersion,
		Package: ccPkg,
	}

	//Resmgmt.Client instance
	resMgmtClient, err := ic.ResourceMgmtClientByOrg(admin, orgID)
	if err != nil {
		return err
	}
	//Invoke InstallCC for each org id
	responses, err := resMgmtClient.InstallCC(req, resmgmt.WithTargets(targets...))
	if err != nil {
		return errors.Errorf("InstallChaincode returned error: %v", err)
	}

	ccIDVersion := ic.chaincodeID + "." + ic.chaincodeVersion

	var errs []error
	for _, resp := range responses {
		if resp.Info == "already installed" {
			fmt.Printf("Chaincode %s already installed on peer: %s.\n", ccIDVersion, resp.Target)
		} else if resp.Status != http.StatusOK {
			errs = append(errs, errors.Errorf("installCC returned error from peer %s: %s", resp.Target, resp.Info))
		} else {
			fmt.Printf("...successfuly installed chaincode %s on peer %s.\n", ccIDVersion, resp.Target)
		}
	}

	if len(errs) > 0 {
		fmt.Printf("\n Errors returned from InstallCC: %v \n", errs)
		return errs[0]
	}

	return nil
}

func (ic *installChaincodeClient) Invoke() ([]byte, error) {
	var lastErr error
	for orgID, peers := range ic.PeersByOrgID() {
		fmt.Printf("Installing chaincode %s on org[%s] peers:\n", ic.chaincodeID, orgID)
		for _, peer := range peers {
			fmt.Printf("-- %s\n", peer.URL())
		}
		err := ic.installChaincode(orgID, peers)
		if err != nil {
			lastErr = err
		}
	}

	return []byte{0x00}, lastErr
}

func (ic *installChaincodeClient) Terminate() {
	ic.CloseSDK()
}
