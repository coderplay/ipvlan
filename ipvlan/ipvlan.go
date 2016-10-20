package ipvlan

import (
	"net"
	"sync"

	"github.com/docker/libnetwork/datastore"
	"github.com/docker/libnetwork/driverapi"
	"github.com/docker/libnetwork/osl"
	"github.com/docker/libnetwork/types"
	"github.com/docker/libnetwork/drivers/remote/api"
)

const (
	vethLen             = 7
	containerVethPrefix = "eth"
	vethPrefix          = "veth"
	ipvlanType          = "ipvlan" // driver type name
	modeL2              = "l2"     // ipvlan mode l2 is the default
	modeL3              = "l3"     // ipvlan L3 mode
	parentOpt           = "parent" // parent interface -o parent
	modeOpt             = "_mode"  // ipvlan mode ux opt suffix

	// add by Min
	gatewayOpt          = "gateway"
	subnetOpt           = "subnet"
)

var driverModeOpt = ipvlanType + modeOpt // mode -o ipvlan_mode

type endpointTable map[string]*endpoint

type networkTable map[string]*network

type driver struct {
	networks networkTable
	sync.Once
	sync.Mutex
	store datastore.DataStore
}

type endpoint struct {
	id       string
	nid      string
	mac      net.HardwareAddr
	addr     *net.IPNet
	addrv6   *net.IPNet
	srcName  string
	dbIndex  uint64
	dbExists bool
}

type network struct {
	id        string
	sbox      osl.Sandbox
	endpoints endpointTable
	driver    *driver
	config    *configuration
	sync.Mutex
}

func NewDriver() (*driver) {
	return &driver{
		networks: networkTable{},
	}
	// TODO: d.initStore(config)
}

func (d *driver) NetworkAllocate(id string, option map[string]string, ipV4Data, ipV6Data []driverapi.IPAMData) (map[string]string, error) {
	return nil, types.NotImplementedErrorf("not implemented")
}

func (d *driver) NetworkFree(id string) error {
	return types.NotImplementedErrorf("not implemented")
}

func (driver *driver) GetCapabilities() (*api.GetCapabilityResponse, error) {
	return &api.GetCapabilityResponse{ Scope: "local"}, nil
}

func (d *driver) EndpointOperInfo(*api.EndpointInfoRequest) (*api.EndpointInfoResponse, error) {
	return &api.EndpointInfoResponse{Value: make(map[string]interface{}, 0)}, nil
}

func (d *driver) Type() string {
	return ipvlanType
}

func (d *driver) ProgramExternalConnectivity(r *api.ProgramExternalConnectivityRequest) error {
	return nil
}

func (d *driver) RevokeExternalConnectivity(r *api.RevokeExternalConnectivityRequest) error {
	return nil
}

// DiscoverNew is a notification for a new discovery event.
func (d *driver) DiscoverNew(r *api.DiscoveryNotification) error {
	return nil
}

// DiscoverDelete is a notification for a discovery delete event
func (d *driver) DiscoverDelete(r *api.DiscoveryNotification) error {
	return nil
}

func (d *driver) EventNotify(etype driverapi.EventType, nid, tableName, key string, value []byte) {
}
