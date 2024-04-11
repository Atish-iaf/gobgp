package main

import (
	"context"
	"fmt"
	"os"

	"gobgp/pkg/bgp"
	"gobgp/pkg/gobgp"
)

func main() {

	// BGP server configuration
	config := bgp.BGPConfig{
		ASN:        65532,
		ListenPort: 49153,
		IPAddress:  "192.168.1.2",
	}

	// Start BGP server
	fmt.Println("====== Starting BGP server ======")
	bgpServer, err := gobgp.NewGobgpServer(context.Background(), config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get BGP server configuration
	fmt.Println("====== Getting BGP server configuration ======")
	gotBgpConfig, err := bgpServer.GetBGPConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(gotBgpConfig)
	}

	// Get BGP Peers
	// fmt.Println("====== Getting BGP Peers ======")
	// peers, err := bgpServer.ListPeers(context.TODO())
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(peers)
	// }

	peer := []*bgp.BGPPeerConfig{
		{
			ASN:        65533,
			ListenPort: 49154,
			IPAddress:  "192.168.1.3",
		},
		{
			ASN:                 65534,
			ListenPort:          49155,
			IPAddress:           "192.168.1.4",
			GracefulRestartTime: 10,
			AuthPassword:        "abc",
		},
	}
	// Add BGP Peers
	fmt.Println("====== Adding BGP Peers ======")
	err = bgpServer.AddBGPPeer(context.TODO(), peer[0])
	if err != nil {
		fmt.Println(err)
	}
	err = bgpServer.AddBGPPeer(context.TODO(), peer[1])
	if err != nil {
		fmt.Println(err)
	}

	//Getting a BGP Peer
	// fmt.Println("====== Getting BGP Peer ======")
	// bgpPeerState, err := bgpServer.GetBGPPeer(context.TODO(), *peer[0])
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(bgpPeerState)
	// }

	// List BGP Peers
	fmt.Println("====== Listing BGP Peers ======")
	peers, err := bgpServer.ListBGPPeers(context.TODO())
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(peers)
	}

	// Update BGP Peer
	// fmt.Println("====== Updating BGP Peer ======")
	// peer[1].GracefulRestartTime = 20
	// peer[1].AuthPassword = "xyz"
	// peer[1].ASN = 65535
	// peer[1].ListenPort = 49156
	// if err = bgpServer.UpdatePeer(context.TODO(), peer[1]); err != nil {
	// 	fmt.Println(err)
	// }

	// Remove BGP Peer
	// fmt.Println("====== Removing BGP Peer ======")
	// if err = bgpServer.RemovePeer(context.TODO(), peer); err != nil {
	// 	fmt.Println(err)
	// }

	// Reset BGP Peer
	// fmt.Println("====== Resetting BGP Peer ======")
	// reserReq := bgp.ResetPeerRequest{
	// 	IPAddress: peer[0].IPAddress,
	// 	SoftReset: true,
	// 	Direction: bgp.ResetDirectionIn,
	// }
	// if err := bgpServer.ResetBGPPeer(context.TODO(), reserReq); err != nil {
	// 	fmt.Println(err)
	// }

	// List BGP Peers
	// fmt.Println("====== Listing BGP Peers ======")
	// peers, err = bgpServer.ListBGPPeers(context.TODO())
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(peers)
	// }

	// Advertise routes
	fmt.Println("====== Advertising routes ======")
	err = bgpServer.AdvertiseRoutes(context.TODO(), []string{})
	if err != nil {
		fmt.Println(err)
	}

	// List advertised routes
	fmt.Println("====== Listing advertised routes ======")
	routes, err := bgpServer.ListAdvertisedRoutes(context.TODO())
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(routes)
	}

	// Stop BGP server
	fmt.Println("====== Stopping BGP server ======")
	bgpServer.StopBGPServer()
}
