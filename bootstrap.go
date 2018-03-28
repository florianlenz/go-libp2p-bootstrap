package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	"gx/ipfs/QmQViVWBHbU6HmYjXcdNq7tVASCNgdg64ZGcauuDkLCivW/go-ipfs-addr"
	"gx/ipfs/QmXauCuJzmzapetmC6W4TuDJLL1yFFrVzSHoWv8YdbmnxH/go-libp2p-peerstore"
	"gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
)

type Bootstrap struct {
	minPeers       int
	bootstrapPeers []*peerstore.PeerInfo
	host           host.Host
	notifiee       *NotifyBundle
	bootstrapping  bool
}

//Get the amount of peer's we are connected to
func (b *Bootstrap) amountConnPeers() int {
	return len(b.host.Network().Peers())
}

//Check if the bootstrapping is locked
func (b *Bootstrap) locked() bool {
	return b.bootstrapping
}

//Lock bootstrapping which will prevent from bootstrapping again
//when the previous bootstrap hasn't finished
func (b *Bootstrap) lock() {

	if b.bootstrapping == true {
		panic("Bootstrapping is already locked")
	}

	b.bootstrapping = true

}

//Unlock bootstrapping so that we can again bootstrap
//in case we have to
func (b *Bootstrap) unlock() {

	if b.bootstrapping == false {
		panic("Bootstrapping is already unlocked")
	}

	b.bootstrapping = false

}

//Start bootstrapping
func (b *Bootstrap) bootstrapp() {

	//Lock bootstrapping
	b.lock()

	go func() {

		for b.amountConnPeers() < b.minPeers {

			c := make(chan struct{})

			for _, v := range b.bootstrapPeers {

				go func() {
					if b.amountConnPeers() < b.minPeers {
						ctx := context.Background()
						if err := b.host.Connect(ctx, *v); err != nil {
							fmt.Println("Failed to connect to: ", v)
							c <- struct{}{}
							return
						}
						fmt.Println("Connected to: ", v)
						c <- struct{}{}
						return
					}
					c <- struct{}{}
				}()

				<-c

			}

		}
		//Unlock bootstrapping
		b.unlock()
	}()
}

//Start bootstrapping
func (b *Bootstrap) Start() {

	//Listener
	notifyBundle := NotifyBundle{
		DisconnectedF: func(network net.Network, conn net.Conn) {
			fmt.Println("Dropped connnection to peer: ", conn.RemotePeer().String())
			//Only bootstrapp when we are currently not bootstrapping
			if b.locked() == false {
				b.bootstrapp()
			}
		},
	}

	//Register listener to react on dropped connections
	b.host.Network().Notify(&notifyBundle)

	b.bootstrapp()
}

//Create new bootstrapper
func NewBootstrap(h host.Host, bootstrapPeers []string, minPeers int) (error, Bootstrap) {

	if minPeers > len(bootstrapPeers) {
		return errors.New(fmt.Sprintf("Too less bootstrapping nodes. Expected at least: %d, got: %d", minPeers, len(bootstrapPeers))), Bootstrap{}
	}

	var peers []*peerstore.PeerInfo

	for _, v := range bootstrapPeers {
		iAddr, err := ipfsaddr.ParseString(v)

		if err != nil {
			return err, Bootstrap{}
		}

		pInfo, err := peerstore.InfoFromP2pAddr(iAddr.Multiaddr())

		if err != nil {
			return err, Bootstrap{}
		}

		peers = append(peers, pInfo)
	}

	return nil, Bootstrap{
		minPeers:       minPeers,
		bootstrapPeers: peers,
		host:           h,
	}

}
