package netutil

import (
	"github.com/gookit/slog"
	"net"
)

type NetInterface struct {
	RawInterface net.Interface
	CurrentIP    net.IP
	MaxIP        net.IP
	Broadcast    bool
}

func GetNetInterfacesByName(name string, array []NetInterface) (*NetInterface, error) {
	for _, netInterface := range array {
		if netInterface.RawInterface.Name == name {
			return &netInterface, nil
		}
	}
	return nil, ERR_NO_INTERFACE_MATCH
}

func GetAvailableNetInterfaces() ([]NetInterface, error) {
	var availableInterfaces []NetInterface
	interfaces, err := net.Interfaces()
	if err != nil {
		slog.Warnf("Failed to read interfaces from sys: %v", err)
		return availableInterfaces, ERR_FAILED_TO_READ_INTERFACE
	}
	for _, face := range interfaces {
		var maxIpFields net.IP
		var curIpFields net.IP
		addrs, err := face.Addrs()
		for _, addr := range addrs {
			if subnetIp, ok := addr.(*net.IPNet); ok {
				ipv4 := subnetIp.IP.To4()
				if ipv4 != nil {
					curIpFields = ipv4
					for i := 0; i < 4; i++ {
						maxIpFields = append(maxIpFields, ipv4[i]|(^subnetIp.Mask[i]))
					}
					break
				}
			}
		}
		broadcast := false
		if (face.Flags & (net.FlagUp | net.FlagLoopback | net.FlagBroadcast)) == (net.FlagBroadcast | net.FlagUp) {
			if err != nil {
				slog.Warnf("Failed to get UDP addrs from interface: %v. Ignored", err)
				continue
			}
			broadcast = true
		}
		availableInterfaces = append(availableInterfaces, NetInterface{RawInterface: face, MaxIP: maxIpFields, CurrentIP: curIpFields, Broadcast: broadcast})
	}
	return availableInterfaces, nil
}
func GetBroadcastInterfaces() ([]NetInterface, error) {
	var broadcastInterface []NetInterface
	interfaces, err := GetAvailableNetInterfaces()
	if err != nil {
		return nil, err
	}
	for _, v := range interfaces {
		if v.Broadcast {
			broadcastInterface = append(broadcastInterface, v)
		}
	}
	return broadcastInterface, nil
}
