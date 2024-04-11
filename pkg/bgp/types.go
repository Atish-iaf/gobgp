package bgp

import (
	"context"
)

type ResetDirection int

const (
	ResetDirectionIn ResetDirection = iota
	ResetDirectionOut
	ResetDirectionBoth
)

type BGPConfig struct {
	// AS number of BGP speaker.
	ASN uint32

	// port on which BGP speaker listens.
	ListenPort int32

	// IP address of BGP speaker.
	IPAddress string
}

type BGPPeerConfig struct {
	// AS number of BGP speaker.
	ASN uint32

	// port on which BGP speaker listens.
	ListenPort int32

	// IP address of BGP speaker.
	IPAddress string

	AuthPassword string

	GracefulRestartTime uint32

	// External or Internal for eBGP and iBGP respectively.
	PeerType string
}

type BGPPeerState struct {
	Config BGPPeerConfig

	// Uptime is time since session got established.
	Uptime int64

	SessionState string
}

type ResetPeerRequest struct {
	IPAddress string

	Direction ResetDirection

	SoftReset bool

	Communication string
}

type Interface interface {
	// StopBGPServer stops BGP server.
	StopBGPServer()

	// AddBGPPeer adds BGP peer.
	AddBGPPeer(ctx context.Context, peer *BGPPeerConfig) error

	// UpdateBGPPeer updates BGP peer.
	UpdateBGPPeer(ctx context.Context, peer *BGPPeerConfig) error

	// RemoveBGPPeer removes BGP peer.
	RemoveBGPPeer(ctx context.Context, peer *BGPPeerConfig) error

	// ResetBGPPeer resets BGP peering with the provided BGP peer IP address.
	ResetBGPPeer(ctx context.Context, resetPeer ResetPeerRequest) error

	// AdvertiseRoutes advertises BGP Path to all configured BGP peers.
	AdvertiseRoutes(ctx context.Context, routes []string) error

	// WithdrawRoutes removes BGP Path from all peers.
	WithdrawRoutes(ctx context.Context, routes []string) error

	// ListAdvertisedRoutes lists all advertised BGP paths.
	ListAdvertisedRoutes(ctx context.Context) (routes []string, err error)

	// ListBGPPeers lists the state of all BGP peers.
	ListBGPPeers(ctx context.Context) ([]BGPPeerState, error)

	// GetBGPConfig returns local BGP speaker configurations from BGP server.
	GetBGPConfig(ctx context.Context) (bgpConfig BGPConfig, err error)

	// GetBGPPeer return existing BGP peer state from BGP server.
	GetBGPPeer(ctx context.Context, peer BGPPeerConfig) (BGPPeerState, error)
}
