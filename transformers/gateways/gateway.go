package gateways

// Gateway is an interface that represents some request/response network transport.
type Gateway interface {
	// Start tells gateway to start listening for incoming requests.
	Start() (err error)

	// Stop tells gateway to stop listening for incoming requests and exit gracefully.
	Stop()
}
