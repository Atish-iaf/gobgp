package gobgp

import (
	"context"
	"fmt"
	"gobgp/pkg/bgp"

	apipb "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/server"
	"google.golang.org/protobuf/types/known/anypb"
)

type GobgpServer struct {
	Server *server.BgpServer
}

// NewGobgpServer starts gobgp server with provided configuration.
func NewGobgpServer(ctx context.Context, bgpConfig bgp.BGPConfig) (bgp.Interface, error) {

	s := server.NewBgpServer()
	go s.Serve()

	startBgpReq := &apipb.StartBgpRequest{
		Global: &apipb.Global{
			Asn:        bgpConfig.ASN,        // optional by gobgp, default is 0
			RouterId:   bgpConfig.IPAddress,  // required by gobgp
			ListenPort: bgpConfig.ListenPort, // optional by gobgp, default is 179
		},
	}

	if err := s.StartBgp(context.Background(), startBgpReq); err != nil {
		return nil, fmt.Errorf("failed to start goBGP server: %w", err)
	}

	return &GobgpServer{
		Server: s,
	}, nil
}

// GetBGPConfig returns the BGP configuration from gobgp server.
func (g *GobgpServer) GetBGPConfig(ctx context.Context) (bgp.BGPConfig, error) {
	getBGPRes, err := g.Server.GetBgp(ctx, &apipb.GetBgpRequest{})
	if err != nil {
		return bgp.BGPConfig{}, fmt.Errorf("failed to get gobgp config from gobgp server: %w", err)
	}

	if getBGPRes.Global == nil {
		return bgp.BGPConfig{}, fmt.Errorf("gobgp server returned nil config")
	}

	goBGPConfig := bgp.BGPConfig{
		ASN:        getBGPRes.Global.Asn,
		IPAddress:  getBGPRes.Global.RouterId,
		ListenPort: getBGPRes.Global.ListenPort,
	}

	return goBGPConfig, nil
}

// StopBGPServer stops gobgp server.
func (g *GobgpServer) StopBGPServer() {
	g.Server.Stop()
	if g.Server != nil {
		fmt.Println("After stopping gobgp server, server is not NIL")
	} else {
		fmt.Println("After stopping gobgp server, server is NIL")
	}
}

func (g *GobgpServer) AdvertiseRoutes(ctx context.Context, routes []string) error {
	nlri, _ := anypb.New(&apipb.IPAddressPrefix{
		Prefix:    "10.0.0.0",
		PrefixLen: 24,
	})
	a1, _ := anypb.New(&apipb.OriginAttribute{
		Origin: 0,
	})
	a2, _ := anypb.New(&apipb.NextHopAttribute{
		NextHop: "10.0.0.1",
	})
	attrs := []*anypb.Any{a1, a2}
	r := &apipb.AddPathRequest{
		TableType: apipb.TableType_ADJ_OUT,
		Path: &apipb.Path{
			Nlri: nlri,
			Family: &apipb.Family{ // family is mandatory
				Afi:  apipb.Family_AFI_IP,
				Safi: apipb.Family_SAFI_UNICAST,
			},
			Pattrs: attrs, // nexthop, origin are mandatory
		},
	}
	res, err := g.Server.AddPath(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to add gobgp path :%w", err)
	}

	fmt.Printf("UIDs %v", res.Uuid)
	return nil
}

func (g *GobgpServer) WithdrawRoutes(ctx context.Context, routes []string) error {
	return nil
}

func (g *GobgpServer) ListAdvertisedRoutes(ctx context.Context) (routes []string, err error) {

	fn := func(destination *apipb.Destination) {
		routes = append(routes, destination.Prefix)
	}
	r := &apipb.ListPathRequest{
		TableType: apipb.TableType_ADJ_OUT,
		Family: &apipb.Family{
			Afi:  apipb.Family_AFI_IP,
			Safi: apipb.Family_SAFI_UNICAST,
		},
		Name: "192.168.1.3",
	}
	if err = g.Server.ListPath(ctx, r, fn); err != nil {
		return []string{}, fmt.Errorf("failed to list gobgp paths :%w", err)
	}

	return routes, nil
}
