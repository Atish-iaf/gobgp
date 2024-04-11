package gobgp

import (
	"context"
	"fmt"
	"gobgp/pkg/bgp"

	apipb "github.com/osrg/gobgp/v3/api"
)

// AddBGPPeer adds a BGP peer with given configuration to gobgp server.
func (g *GobgpServer) AddBGPPeer(ctx context.Context, peer *bgp.BGPPeerConfig) error {
	p := &apipb.AddPeerRequest{
		Peer: &apipb.Peer{
			Conf: &apipb.PeerConf{
				PeerAsn: peer.ASN,
				//Type:            apipb.PeerType_EXTERNAL,
				NeighborAddress: peer.IPAddress,
				AuthPassword:    peer.AuthPassword,
			},
			Transport: &apipb.Transport{
				RemotePort: uint32(peer.ListenPort),
			},
			GracefulRestart: &apipb.GracefulRestart{
				Enabled: true,
			},
		},
	}

	if err := g.Server.AddPeer(ctx, p); err != nil {
		return fmt.Errorf("failed to add gobgp peer: %w", err)
	}

	return nil
}

func (g *GobgpServer) UpdateBGPPeer(ctx context.Context, peer *bgp.BGPPeerConfig) error {
	existingGobgpPeer, err := g.getExistingGobgpPeer(ctx, peer.IPAddress)
	if err != nil {
		return err
	}

	newGobgpPeer := &apipb.Peer{
		Conf:            existingGobgpPeer.Conf,
		Transport:       existingGobgpPeer.Transport,
		GracefulRestart: existingGobgpPeer.GracefulRestart,
	}
	newGobgpPeer.Conf.PeerAsn = peer.ASN
	newGobgpPeer.Transport.RemotePort = uint32(peer.ListenPort)
	newGobgpPeer.Conf.AuthPassword = peer.AuthPassword
	newGobgpPeer.GracefulRestart.RestartTime = peer.GracefulRestartTime

	updateRes, err := g.Server.UpdatePeer(ctx, &apipb.UpdatePeerRequest{Peer: newGobgpPeer})
	if err != nil {
		return fmt.Errorf("failed while updating peer %v:%v with ASN %v: %w", newGobgpPeer.Conf.NeighborAddress, newGobgpPeer.Transport.RemotePort, newGobgpPeer.Conf.PeerAsn, err)
	}

	resetReq := &apipb.ResetPeerRequest{
		Address: newGobgpPeer.Conf.NeighborAddress,
	}

	// In which case softReset is needed ?
	if updateRes.NeedsSoftResetIn {
		resetReq.Soft = true
		resetReq.Direction = apipb.ResetPeerRequest_IN
	}
	if err = g.Server.ResetPeer(ctx, resetReq); err != nil {
		return fmt.Errorf("failed while resetting peer %v:%v with ASN %v: %w", newGobgpPeer.Conf.NeighborAddress, newGobgpPeer.Transport.RemotePort, newGobgpPeer.Conf.PeerAsn, err)
	}
	return nil
}

// RemoveBGPPeer removes BGP peer from gobgp server.
func (g *GobgpServer) RemoveBGPPeer(ctx context.Context, peer *bgp.BGPPeerConfig) error {
	peerReq := &apipb.DeletePeerRequest{
		Address: peer.IPAddress,
	}

	if err := g.Server.DeletePeer(ctx, peerReq); err != nil {
		return fmt.Errorf("failed to remove bgp peer %v %v: %w", peer.ASN, peer.IPAddress, err)
	}

	return nil
}

// ListBGPPeers returns the state of all BGP peers from gobgp server.
func (g *GobgpServer) ListBGPPeers(ctx context.Context) ([]bgp.BGPPeerState, error) {
	var peers []bgp.BGPPeerState
	fn := func(peer *apipb.Peer) {
		if peer == nil {
			return
		}
		peerState := toBGPPeerState(peer)
		peers = append(peers, peerState)
	}

	err := g.Server.ListPeer(ctx, &apipb.ListPeerRequest{}, fn)
	if err != nil {
		return []bgp.BGPPeerState{}, fmt.Errorf("failed to list gobgp peers: %w", err)
	}
	return peers, nil
}

func (g *GobgpServer) getExistingGobgpPeer(ctx context.Context, address string) (*apipb.Peer, error) {
	var existingGobgpPeer *apipb.Peer
	fn := func(gobgpPeer *apipb.Peer) {
		existingGobgpPeer = gobgpPeer
	}
	err := g.Server.ListPeer(ctx, &apipb.ListPeerRequest{
		Address: address,
	}, fn)
	if err != nil {
		return existingGobgpPeer, fmt.Errorf("listing gobgp peer failed: %w", err)
	}
	if existingGobgpPeer == nil {
		return existingGobgpPeer, fmt.Errorf("could not find existing gobgp peer IP: %s", address)
	}
	return existingGobgpPeer, nil
}

func (g *GobgpServer) ResetBGPPeer(ctx context.Context, resetPeer bgp.ResetPeerRequest) error {
	r := &apipb.ResetPeerRequest{
		Address:       resetPeer.IPAddress, // "" or "all", will reset all peers
		Communication: resetPeer.Communication,
		Soft:          resetPeer.SoftReset,
		Direction:     toGobgpSoftResetDirection(resetPeer.Direction),
	}
	if err := g.Server.ResetPeer(ctx, r); err != nil {
		return fmt.Errorf("failed to reset gobgp peer: %w", err)
	}
	return nil
}

func toGobgpSoftResetDirection(direction bgp.ResetDirection) apipb.ResetPeerRequest_SoftResetDirection {
	switch direction {
	case bgp.ResetDirectionIn:
		return apipb.ResetPeerRequest_IN
	case bgp.ResetDirectionOut:
		return apipb.ResetPeerRequest_OUT
	}
	return apipb.ResetPeerRequest_BOTH
}

func (g *GobgpServer) GetBGPPeer(ctx context.Context, peer bgp.BGPPeerConfig) (bgp.BGPPeerState, error) {
	var bgpPeerState bgp.BGPPeerState
	if peer.IPAddress == "" {
		return bgpPeerState, fmt.Errorf("cannot get gobgp peer with empty IP address")
	}

	gobgpPeer, err := g.getExistingGobgpPeer(ctx, peer.IPAddress)
	if err != nil {
		return bgpPeerState, fmt.Errorf("failed to get gobgp peer: %s : %w", peer.IPAddress, err)
	}
	bgpPeerState = toBGPPeerState(gobgpPeer)

	return bgpPeerState, nil
}

func toBGPPeerState(gobgpPeer *apipb.Peer) bgp.BGPPeerState {
	var bgpPeerState bgp.BGPPeerState
	if gobgpPeer.Conf != nil {
		bgpPeerState.Config.ASN = gobgpPeer.Conf.PeerAsn
		bgpPeerState.Config.IPAddress = gobgpPeer.Conf.NeighborAddress
		bgpPeerState.Config.ListenPort = int32(gobgpPeer.Transport.RemotePort)
		bgpPeerState.Config.PeerType = gobgpPeer.Conf.Type.String()
		bgpPeerState.Config.AuthPassword = gobgpPeer.Conf.AuthPassword
	}

	if gobgpPeer.GracefulRestart != nil && gobgpPeer.GracefulRestart.Enabled {
		bgpPeerState.Config.GracefulRestartTime = gobgpPeer.GracefulRestart.RestartTime
	}

	if gobgpPeer.State != nil {
		bgpPeerState.SessionState = gobgpPeer.State.SessionState.String()
	}

	return bgpPeerState
}
