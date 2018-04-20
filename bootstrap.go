package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/ipfs/go-log"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
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
	minPeers                int
	bootstrapPeers          []*peerstore.PeerInfo
	host                    host.Host
	notifiee                *net.NotifyBundle
	interfaceListenerLocked bool
	bootstrapInterval       time.Duration
	hardBootstrap           time.Duration
	started                 bool
}

//Lock the interface listener
func (b *Bootstrap) lockInterfaceListener() {

	if b.interfaceListenerLocked == true {
		panic("Interface listener is already locked")
	}

	b.interfaceListenerLocked = true
}

//Unlock the interface listener
func (b *Bootstrap) unlockInterfaceListener() {

	if b.interfaceListenerLocked == false {
		panic("Interface listener is already unlocked")
	}

	b.interfaceListenerLocked = false
}

//Is the interface listener locked
func (b *Bootstrap) isInterfaceListenerLocked() bool {
	return b.interfaceListenerLocked
}

//Get the amount of peer's we are connected to
func (b *Bootstrap) amountConnPeers() int {
	logger.Debug("Amount of connected peer's: ", len(b.host.Network().Peers()))
	return len(b.host.Network().Peers())
}

//Register a network state change handler
func (b *Bootstrap) networkInterfaceListener() {

	//Only register listener when we are connected
	//to too less peer's
	if b.amountConnPeers() >= b.minPeers {
		return
	}

	//Lock down the interface listener
	//to prevent a second listener registration
	b.lockInterfaceListener()

	//Register latest network state
	mas, err := b.host.Network().InterfaceListenAddresses()
	if err != nil {
		panic(err)
	}
	lastNetworkState := len(mas)
	now := time.Now()
	logger.Debug("Addresses at install time: ", lastNetworkState)

	go func() {

		for {

			//In case bootstrapping is stopped
			//we want to exit
			if b.started == false {
				break
			}

			//After x we want to do a hard bootstrap.
			//Hard bootstrap mean's that we bypass the
			//check for an delta on the addresses and try to bootstrap
			if time.Now().After(now.Add(b.hardBootstrap)) {
				logger.Debug("Hard bootstrap")
				b.Bootstrap()
				now = time.Now()
			}

			//Get current network state
			mas, err := b.host.Network().InterfaceListenAddresses()
			if err != nil {
				panic(err)
			}

			//Bootstrap on network delta (delta between the amount of addresses)
			if len(mas) != lastNetworkState {
				lastNetworkState = len(mas)
				b.Bootstrap()
			}

			//We can un register the handler when we are connected to enough peer's
			if len(b.host.Network().Peers()) >= b.minPeers {
				break
			}

			//Pause before we continue with bootstrap attempts
			time.Sleep(b.bootstrapInterval)

			logger.Debug("Next listener round after: ", b.bootstrapInterval)
		}

		//Time to unlock the interface listener
		//since the for loop will only be "done"
		//when we are connected to enough peer's
		logger.Debug("Free network interface listener")
		b.unlockInterfaceListener()

	}()

}

//Start bootstrapping
func (b *Bootstrap) Bootstrap() error {

	if !b.started {
		return errors.New("you need to to call Start() first in order to manually bootstrap")
	}

	c := make(chan struct{})

	var e error

	for _, v := range b.bootstrapPeers {

		go func() {
			if b.amountConnPeers() < b.minPeers {
				ctx := context.Background()
				err := b.host.Connect(ctx, *v)
				if err != nil {
					logger.Debug("Failed to connect to peer: ", v)
					e = err
					c <- struct{}{}
					return
				}
				logger.Debug("Connected to: ", v)
				c <- struct{}{}
				return
			}
			c <- struct{}{}
		}()

		<-c

	}

	return e

}

//Stop the bootstrap service
func (b *Bootstrap) Stop() error {
	if b.started == false {
		return errors.New("bootstrap must be started in order to stop it")
	}

	b.host.Network().StopNotify(b.notifiee)
	b.started = false
	return nil
}

//Start bootstrapping
func (b *Bootstrap) Start() error {

	if b.started == true {
		return errors.New("already started")
	}
	b.started = true

	//Listener
	notifyBundle := net.NotifyBundle{
		DisconnectedF: func(network net.Network, conn net.Conn) {
			logger.Debug("Dropped connection to peer: ", conn.RemotePeer())
			if b.isInterfaceListenerLocked() == false {
				logger.Debug("Install network interface listener")
				b.networkInterfaceListener()
			}
		},
	}

	//Register listener to react on dropped connections
	b.host.Network().Notify(&notifyBundle)

	if errors := b.Bootstrap(); len(errors) != 0 {
		//In case we fail to start,
		//Register network interface listener
		b.networkInterfaceListener()
	}

	return nil
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
		minPeers:                c.MinPeers,
		bootstrapPeers:          peers,
		host:                    h,
		hardBootstrap:           c.HardBootstrap,
		bootstrapInterval:       c.BootstrapInterval,
		interfaceListenerLocked: false,
		started:                 false,
	}

}
