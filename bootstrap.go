package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	peerState "github.com/florianlenz/go-libp2p-bootstrap/state/peers"
	startedState "github.com/florianlenz/go-libp2p-bootstrap/state/started"
	log "gx/ipfs/QmTG23dvpBCBjqQwyDxV8CQT6jmS4PSftNr1VqHhE3MLy7/go-log"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	net "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	peerstore "gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	host "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var logger = log.Logger("bootstrap")

//Bootstrap configuration
//"HardBootstrap" is the time after we
//dial to peer's in order to prove if we are connected ot the WWW
//instead of waiting for a delta in our addresses.
//This shouldn't be done too often since it can lead to problems
//(https://github.com/libp2p/go-libp2p-swarm/issues/37).
//You can chose something that is higher than one minute.
type Config struct {
	BootstrapPeers    []string
	MinPeers          int
	BootstrapInterval time.Duration
	HardBootstrap     time.Duration
}

type Bootstrap struct {
	minPeers          int
	bootstrapPeers    []*peerstore.PeerInfo
	host              host.Host
	notifiee          *net.NotifyBundle
	bootstrapInterval time.Duration
	hardBootstrap     time.Duration
	startedState      *startedState.State
	peerState         *peerState.State
}

//Bootstrap thought the list of bootstrap peer's
func (b *Bootstrap) Bootstrap(ctx context.Context) error {

	if !b.startedState.HasStarted() {
		return errors.New("you need to to call Start() first in order to manually bootstrap")
	}

	var e error

	var wg sync.WaitGroup

	for _, peer := range b.bootstrapPeers {

		wg.Add(1)
		go func(peer *peerstore.PeerInfo) {
			defer wg.Done()
			if b.peerState.Amount() < b.minPeers {
				if err := b.host.Connect(ctx, *peer); err != nil {
					logger.Debug("Failed to connect to peer: ", peer)
					e = err
					return
				}
				logger.Debug("Connected to: ", peer)
			}
		}(peer)

	}

	wg.Wait()

	return e

}

//Stop the bootstrap service
func (b *Bootstrap) Close() error {
	if !b.startedState.HasStarted() {
		return errors.New("bootstrap must be started in order to stop it")
	}

	b.host.Network().StopNotify(b.notifiee)
	b.startedState.Stop()
	return nil
}

//Start bootstrapping
func (b *Bootstrap) Start(ctx context.Context) error {

	//Pre start conditions
	if b.startedState.HasStarted() {
		return errors.New("already started")
	}
	b.startedState.Start()

	//Set initial amount of peer's
	b.peerState.SetAmountOfPeers(len(b.host.Network().Peers()))

	//Listener that updates the amount of connected peer's
	notifyBundle := net.NotifyBundle{
		DisconnectedF: func(network net.Network, conn net.Conn) {
			b.peerState.SetAmountOfPeers(len(network.Peers()))
		},
		ConnectedF: func(network net.Network, conn net.Conn) {
			b.peerState.SetAmountOfPeers(len(network.Peers()))
		},
	}
	b.host.Network().Notify(&notifyBundle)

	//Do an initial bootstrap
	err := b.Bootstrap(ctx)

	//Start the worker
	go func() {

		lastNetworkState := len(b.host.Network().Peers())

		hb := time.Now()

		for {

			//Break in case we stopped the bootstrapping process
			if !b.startedState.HasStarted() {
				logger.Warning("stop worker since bootstrap hasn't started")
				break
			}

			myAddresses := len(b.host.Network().Peers())
			connectedPeers := b.peerState.Amount()

			//Continue when we are connected to the minPeer amount
			if connectedPeers >= b.minPeers {
				logger.Info("already connected to enough peer's")
				hb = time.Now()
				time.Sleep(b.bootstrapInterval)
				continue
			}

			//Bootstrap on network delta (delta between the amount of our addresses)
			if myAddresses != lastNetworkState {
				lastNetworkState = myAddresses
				b.Bootstrap(context.Background())
			}

			//Hard bootstrap
			if time.Now().After(hb.Add(b.hardBootstrap)) {
				logger.Debug("Hard bootstrap")
				b.Bootstrap(context.Background())
				hb = time.Now()
			}

			//Wait some time for the new round
			time.Sleep(b.bootstrapInterval)

		}

	}()

	return err

}

//Create new bootstrap service
func NewBootstrap(h host.Host, c Config) (error, *Bootstrap) {

	if c.MinPeers > len(c.BootstrapPeers) {
		return errors.New(fmt.Sprintf("Too less bootstrapping nodes. Expected at least: %d, got: %d", c.MinPeers, len(c.BootstrapPeers))), &Bootstrap{}
	}

	var peers []*peerstore.PeerInfo

	for _, v := range c.BootstrapPeers {
		addr, err := ma.NewMultiaddr(v)

		if err != nil {
			return err, &Bootstrap{}
		}

		pInfo, err := peerstore.InfoFromP2pAddr(addr)

		if err != nil {
			return err, &Bootstrap{}
		}

		peers = append(peers, pInfo)
	}

	return nil, &Bootstrap{
		minPeers:          c.MinPeers,
		bootstrapPeers:    peers,
		host:              h,
		hardBootstrap:     c.HardBootstrap,
		bootstrapInterval: c.BootstrapInterval,
		startedState:      startedState.StateFactory(),
		peerState:         peerState.StateFactory(),
	}

}
