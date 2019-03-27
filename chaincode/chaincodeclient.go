package chaincode

//ChaincodeClient defines methods for interacting with the chaincode on the fabric network
type ChaincodeClient interface {
	Invoke() ([]byte, error)
	Terminate()
}
