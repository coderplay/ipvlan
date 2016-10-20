package ipvlan

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libnetwork/netlabel"
	"github.com/docker/libnetwork/ns"
	"github.com/docker/libnetwork/osl"
	"github.com/docker/libnetwork/types"
	"github.com/docker/libnetwork/drivers/remote/api"
	"net"
)

// CreateEndpoint assigns the mac, ip and endpoint id for the new container
func (d *driver) CreateEndpoint(r *api.CreateEndpointRequest) (*api.CreateEndpointResponse, error) {
	defer osl.InitOSContext()()

	if err := validateID(r.NetworkID, r.EndpointID); err != nil {
		return nil, err
	}
	n, err := d.getNetwork(r.NetworkID)
	if err != nil {
		return nil, fmt.Errorf("network id %q not found", r.NetworkID)
	}

	if len(r.Interface.MacAddress) != 0 {
		return nil, fmt.Errorf("%s interfaces do not support custom mac address assigment", ipvlanType)
	}

	ep := &endpoint{
		id:     r.NetworkID,
		nid:    r.EndpointID,

	}

	if len(r.Interface.Address) == 0 {
		return nil, fmt.Errorf("create endpoint was not passed an IP address")
	}
	if len(r.Interface.Address) > 0 {
		_, addressIPv4, err := net.ParseCIDR(r.Interface.Address)
		if err != nil {
			return nil, fmt.Errorf("%s is an invalid ipv4 address", r.Interface.Address)
		}
		ep.addr = addressIPv4
	}
	if len(r.Interface.AddressIPv6) > 0 {
		_, addressIPv6, err := net.ParseCIDR(r.Interface.AddressIPv6)
		if err != nil {
			return nil, fmt.Errorf("%s %d is an invalid ipv6 address", r.Interface.AddressIPv6, len(r.Interface.AddressIPv6))
		}
		ep.addrv6 = addressIPv6
	}

	// disallow port mapping -p
	if opt, ok := r.Options[netlabel.PortMap]; ok {
		if _, ok := opt.([]types.PortBinding); ok {
			if len(opt.([]types.PortBinding)) > 0 {
				logrus.Warnf("%s driver does not support port mappings", ipvlanType)
			}
		}
	}
	// disallow port exposure --expose
	if opt, ok := r.Options[netlabel.ExposedPorts]; ok {
		if _, ok := opt.([]types.TransportPort); ok {
			if len(opt.([]types.TransportPort)) > 0 {
				logrus.Warnf("%s driver does not support port exposures", ipvlanType)
			}
		}
	}

	if err := d.storeUpdate(ep); err != nil {
		return nil, fmt.Errorf("failed to save ipvlan endpoint %s to store: %v", ep.id[0:7], err)
	}

	n.addEndpoint(ep)

	resp := &api.CreateEndpointResponse{}
	return resp, nil
}

// DeleteEndpoint remove the endpoint and associated netlink interface
func (d *driver) DeleteEndpoint(r *api.DeleteEndpointRequest) error {
	defer osl.InitOSContext()()
	if err := validateID(r.NetworkID, r.EndpointID); err != nil {
		return err
	}
	n := d.network(r.NetworkID)
	if n == nil {
		return fmt.Errorf("network id %q not found", r.NetworkID)
	}
	ep := n.endpoint(r.EndpointID)
	if ep == nil {
		return fmt.Errorf("endpoint id %q not found", r.EndpointID)
	}
	if link, err := ns.NlHandle().LinkByName(ep.srcName); err == nil {
		ns.NlHandle().LinkDel(link)
	}

	if err := d.storeDelete(ep); err != nil {
		logrus.Warnf("Failed to remove ipvlan endpoint %s from store: %v", ep.id[0:7], err)
	}
	n.deleteEndpoint(ep.id)
	return nil
}
