package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/safchain/ethtool"
)

type InterfaceInfo struct {
	Interface string
	IPs       []string
}

func main() {
	// Start HTTP server in a goroutine
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metric_txt := get_metrics()
		fmt.Fprint(w, metric_txt)
	})

	fmt.Println("Starting HTTP server on :2112")
	if err := http.ListenAndServe("0.0.0.0:2112", nil); err != nil {
		fmt.Printf("Error starting HTTP server: %v\n", err)
	}
}

func get_metrics() string {
	metric_txt := ""
	interfaces := getNetworkInterfaces()
	for _, iface := range interfaces {
		stats, err := getEthtoolInfo(iface.Interface)
		if err != nil {
			fmt.Printf("Error getting ethtool info for %s: %v\n", iface, err)
			continue
		}
		for _, ip := range iface.IPs {
			for statName, statValue := range stats {
				metric_txt += fmt.Sprintf("aws_ec2_network{interface=\"%s\", ip=\"%s\", stat=\"%s\"} %d\n",
					iface.Interface, ip, statName, statValue)
			}
		}
	}
	return metric_txt
}

func getNetworkInterfaces() []InterfaceInfo {
	var ifaces []InterfaceInfo
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error getting network interfaces: %v\n", err)
		return nil
	}

	for _, iface := range interfaces {
		// Skip loopback interface
		if iface.Name == "lo" {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Check if interface has IPv4 address
		hasIPv4 := false
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() != nil {
				hasIPv4 = true
				break
			}
		}

		// Skip if no IPv4 address
		if !hasIPv4 {
			continue
		}

		var ips []string
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}

		ifaces = append(ifaces, InterfaceInfo{
			Interface: iface.Name,
			IPs:       ips,
		})
	}

	return ifaces
}

func getEthtoolInfo(ifaceName string) (map[string]uint64, error) {
	// Create new ethtool handle
	eth, err := ethtool.NewEthtool()
	if err != nil {
		fmt.Printf("Error creating ethtool handle: %v\n", err)
		return nil, err
	}
	defer eth.Close()

	// Get interface statistics using ethtool
	stats, err := eth.Stats(ifaceName)
	if err != nil {
		fmt.Printf("  Error getting ethtool stats: %v\n", err)
		return nil, err
	}

	return stats, nil
}
