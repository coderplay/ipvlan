package ipvlan

import (
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libnetwork/netutils"
	"github.com/docker/libnetwork/ns"
	"github.com/docker/libnetwork/osl"
	"github.com/docker/libnetwork/types"
	"github.com/docker/libnetwork/drivers/remote/api"
)

var (
	defaultV4Route = &api.StaticRoute{
		"0.0.0.0/0",
		types.CONNECTED,
		"0.0.0.0",
	}
	defaultV6Route = &api.StaticRoute{
		"::/0",
		types.CONNECTED,
		"::",
	}
)

// Join method is invoked when a Sandbox is attached to an endpoint.
func (d *driver) Join(r *api.JoinRequest) (*api.JoinResponse, error) {
	defer osl.InitOSContext()()
	n, err := d.getNetwork(r.NetworkID)
	if err != nil {
		return nil,err
	}
	endpoint := n.endpoint(r.EndpointID)
	if endpoint == nil {
		return nil, fmt.Errorf("could not find endpoint with id %s", r.EndpointID)
	}
	// generate a name for the iface that will be renamed to eth0 in the sbox
	containerIfName, err := netutils.GenerateIfaceName(ns.NlHandle(), vethPrefix, vethLen)
	if err != nil {
		return nil, fmt.Errorf("error generating an interface name: %v", err)
	}
	// create the netlink ipvlan interface
	vethName, err := createIPVlan(containerIfName, n.config.Parent, n.config.IpvlanMode)
	if err != nil {
		return nil, err
	}
	// bind the generated iface name to the endpoint
	endpoint.srcName = vethName
	ep := n.endpoint(r.EndpointID)
	if ep == nil {
		return nil, fmt.Errorf("could not find endpoint with id %s", r.EndpointID)
	}

	response := &api.JoinResponse{
		InterfaceName: &api.InterfaceName{
			SrcName:   vethName,
			DstPrefix: containerVethPrefix,
		},
	}

	if n.config.IpvlanMode == modeL3 {
		// disable gateway services to add a default gw using dev eth0 only
		//jinfo.DisableGatewayService()
		response.StaticRoutes = append(response.StaticRoutes, defaultV4Route)
		logrus.Debugf("Ipvlan Endpoint Joined with IPv4_Addr: %s, Ipvlan_Mode: %s, Parent: %s",
			ep.addr.IP.String(), n.config.IpvlanMode, n.config.Parent)
		// If the endpoint has a v6 address, set a v6 default route
		if ep.addrv6 != nil {
			response.StaticRoutes = append(response.StaticRoutes, defaultV6Route)
			logrus.Debugf("Ipvlan Endpoint Joined with IPv6_Addr: %s, Ipvlan_Mode: %s, Parent: %s",
				ep.addrv6.IP.String(), n.config.IpvlanMode, n.config.Parent)
		}
	}
	if n.config.IpvlanMode == modeL2 {
		// parse and correlate the endpoint v4 address with the available v4 subnets
		if len(n.config.Ipv4Subnets) > 0 {
			s := n.getSubnetforIPv4(ep.addr)
			if s == nil {
				return nil, fmt.Errorf("could not find a valid ipv4 subnet for endpoint %s", r.EndpointID)
			}
			response.Gateway = s.GwIP
			logrus.Debugf("Ipvlan Endpoint Joined with IPv4_Addr: %s, Gateway: %s, Ipvlan_Mode: %s, Parent: %s",
				ep.addr.IP.String(), s.GwIP, n.config.IpvlanMode, n.config.Parent)
		}
		// parse and correlate the endpoint v6 address with the available v6 subnets
		if len(n.config.Ipv6Subnets) > 0 {
			s := n.getSubnetforIPv6(ep.addrv6)
			if s == nil {
				return nil, fmt.Errorf("could not find a valid ipv6 subnet for endpoint %s", r.EndpointID)
			}
			response.GatewayIPv6 = s.GwIP
			logrus.Debugf("Ipvlan Endpoint Joined with IPv6_Addr: %s, Gateway: %s, Ipvlan_Mode: %s, Parent: %s",
				ep.addrv6.IP.String(), s.GwIP, n.config.IpvlanMode, n.config.Parent)
		}
	}

	if err = d.storeUpdate(ep); err != nil {
		return nil, fmt.Errorf("failed to save ipvlan endpoint %s to store: %v", ep.id[0:7], err)
	}

	return response, nil
}

// Leave method is invoked when a Sandbox detaches from an endpoint.
func (d *driver) Leave(r *api.LeaveRequest) error {
	defer osl.InitOSContext()()
	network, err := d.getNetwork(r.NetworkID)
	if err != nil {
		return err
	}
	endpoint, err := network.getEndpoint(r.EndpointID)
	if err != nil {
		return err
	}
	if endpoint == nil {
		return fmt.Errorf("could not find endpoint with id %s", r.EndpointID)
	}

	return nil
}

// getSubnetforIPv4 returns the ipv4 subnet to which the given IP belongs
func (n *network) getSubnetforIPv4(ip *net.IPNet) *ipv4Subnet {
	for _, s := range n.config.Ipv4Subnets {
		_, snet, err := net.ParseCIDR(s.SubnetIP)
		if err != nil {
			return nil
		}
		// first check if the mask lengths are the same
		i, _ := snet.Mask.Size()
		j, _ := ip.Mask.Size()
		if i != j {
			continue
		}
		if snet.Contains(ip.IP) {
			return s
		}
	}

	return nil
}

// getSubnetforIPv6 returns the ipv6 subnet to which the given IP belongs
func (n *network) getSubnetforIPv6(ip *net.IPNet) *ipv6Subnet {
	for _, s := range n.config.Ipv6Subnets {
		_, snet, err := net.ParseCIDR(s.SubnetIP)
		if err != nil {
			return nil
		}
		// first check if the mask lengths are the same
		i, _ := snet.Mask.Size()
		j, _ := ip.Mask.Size()
		if i != j {
			continue
		}
		if snet.Contains(ip.IP) {
			return s
		}
	}

	return nil
}
