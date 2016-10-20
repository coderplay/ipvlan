package ipvlan

import (
	"log"
	"net/http"

	"github.com/docker/go-plugins-helpers/sdk"
	"github.com/docker/libnetwork/drivers/remote/api"
)

const (
	manifest = `{"Implements": ["NetworkDriver"]}`
	// LocalScope is the correct scope response for a local scope driver
	LocalScope = `local`
	// GlobalScope is the correct scope response for a global scope driver
	GlobalScope = `global`

	capabilitiesPath   = "/NetworkDriver.GetCapabilities"
	createNetworkPath  = "/NetworkDriver.CreateNetwork"
	deleteNetworkPath  = "/NetworkDriver.DeleteNetwork"
	createEndpointPath = "/NetworkDriver.CreateEndpoint"
	endpointInfoPath   = "/NetworkDriver.EndpointOperInfo"
	deleteEndpointPath = "/NetworkDriver.DeleteEndpoint"
	joinPath           = "/NetworkDriver.Join"
	leavePath          = "/NetworkDriver.Leave"
	discoverNewPath    = "/NetworkDriver.DiscoverNew"
	discoverDeletePath = "/NetworkDriver.DiscoverDelete"
	programExtConnPath = "/NetworkDriver.ProgramExternalConnectivity"
	revokeExtConnPath  = "/NetworkDriver.RevokeExternalConnectivity"
)

// Driver represent the interface a driver must fulfill.
type Driver interface {
	GetCapabilities() (*api.GetCapabilityResponse, error)
	CreateNetwork(*api.CreateNetworkRequest) error
	DeleteNetwork(*api.DeleteNetworkRequest) error
	CreateEndpoint(*api.CreateEndpointRequest) (*api.CreateEndpointResponse, error)
	DeleteEndpoint(*api.DeleteEndpointRequest) error
	EndpointOperInfo(*api.EndpointInfoRequest) (*api.EndpointInfoResponse, error)
	Join(*api.JoinRequest) (response *api.JoinResponse, error error)
	Leave(*api.LeaveRequest) error
	DiscoverNew(*api.DiscoveryNotification) error
	DiscoverDelete(*api.DiscoveryNotification) error
	ProgramExternalConnectivity(*api.ProgramExternalConnectivityRequest) error
	RevokeExternalConnectivity(*api.RevokeExternalConnectivityRequest) error
}

// ErrorResponse is a formatted error message that libnetwork can understand
type ErrorResponse struct {
	Err string
}

// NewErrorResponse creates an ErrorResponse with the provided message
func NewErrorResponse(msg string) *ErrorResponse {
	return &ErrorResponse{Err: msg}
}

// Handler forwards requests and responses between the docker daemon and the plugin.
type Handler struct {
	driver Driver
	sdk.Handler
}

// NewHandler initializes the request handler with a driver implementation.
func NewHandler(driver Driver) *Handler {
	h := &Handler{driver, sdk.NewHandler(manifest)}
	h.initMux()
	return h
}

func (h *Handler) initMux() {
	h.HandleFunc(capabilitiesPath, func(w http.ResponseWriter, r *http.Request) {
		res, err := h.driver.GetCapabilities()
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		if res == nil {
			msg := "Network driver must implement GetCapabilities"
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, res, "")
	})
	h.HandleFunc(createNetworkPath, func(w http.ResponseWriter, r *http.Request) {
		log.Println("Entering go-plugins-helpers createnetwork")
		req := &api.CreateNetworkRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		err = h.driver.CreateNetwork(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, make(map[string]string), "")
	})
	h.HandleFunc(deleteNetworkPath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.DeleteNetworkRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		err = h.driver.DeleteNetwork(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, make(map[string]string), "")
	})
	h.HandleFunc(createEndpointPath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.CreateEndpointRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		res, err := h.driver.CreateEndpoint(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
		}
		sdk.EncodeResponse(w, res, "")
	})
	h.HandleFunc(deleteEndpointPath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.DeleteEndpointRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		err = h.driver.DeleteEndpoint(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, make(map[string]string), "")
	})
	h.HandleFunc(endpointInfoPath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.EndpointInfoRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		res, err := h.driver.EndpointOperInfo(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
		}
		sdk.EncodeResponse(w, res, "")
	})
	h.HandleFunc(joinPath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.JoinRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		res, err := h.driver.Join(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
		}
		sdk.EncodeResponse(w, res, "")
	})
	h.HandleFunc(leavePath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.LeaveRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		err = h.driver.Leave(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, make(map[string]string), "")
	})
	h.HandleFunc(discoverNewPath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.DiscoveryNotification{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		err = h.driver.DiscoverNew(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, make(map[string]string), "")
	})
	h.HandleFunc(discoverDeletePath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.DiscoveryNotification{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		err = h.driver.DiscoverDelete(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, make(map[string]string), "")
	})
	h.HandleFunc(programExtConnPath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.ProgramExternalConnectivityRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		err = h.driver.ProgramExternalConnectivity(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, make(map[string]string), "")
	})
	h.HandleFunc(revokeExtConnPath, func(w http.ResponseWriter, r *http.Request) {
		req := &api.RevokeExternalConnectivityRequest{}
		err := sdk.DecodeRequest(w, r, req)
		if err != nil {
			return
		}
		err = h.driver.RevokeExternalConnectivity(req)
		if err != nil {
			msg := err.Error()
			sdk.EncodeResponse(w, NewErrorResponse(msg), msg)
			return
		}
		sdk.EncodeResponse(w, make(map[string]string), "")
	})
}
