package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	"gx/ipfs/QmQViVWBHbU6HmYjXcdNq7tVASCNgdg64ZGcauuDkLCivW/go-ipfs-addr"
	"gx/ipfs/QmXauCuJzmzapetmC6W4TuDJLL1yFFrVzSHoWv8YdbmnxH/go-libp2p-peerstore"
	"gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	"sync"
)

type Bootstrap struct {
	minPeers       int
	bootstrapPeers []string
	host           host.Host
	notifiee       *NotifyBundle
}

//Bootstrap till we have the required amount of peer's
func (b *Bootstrap) all() {

	wg := sync.WaitGroup{}

	//Loop the nodes
	for _, peer := range b.bootstrapPeers {

		ctx := context.Background()
		iAddr, err := ipfsaddr.ParseString(peer)

		//@todo change this later on
		if err != nil {
			panic(err)
		}

		//Get peer info
		pInfo, err := peerstore.InfoFromP2pAddr(iAddr.Multiaddr())

		//Connect to the peer
		wg.Add(1)
		go func() {

			if len(b.host.Peerstore().Peers()) < b.minPeers {
				if err := b.host.Connect(ctx, *pInfo); err != nil {
					panic(err)
				}
				fmt.Println("connected to: ", pInfo)
				wg.Done()
				return
			}
			wg.Done()
		}()

	}

	wg.Wait()

}

//Start bootstrapping
func (b *Bootstrap) Start() {

	//Listener
	notifyBundle := NotifyBundle{
		DisconnectedF: func(network net.Network, conn net.Conn) {

			b.all()

		},
	}

	//Register listener
	b.host.Network().Notify(&notifyBundle)

	//Initial bootstrap
	b.all()
}

//Create new bootstrapper
func NewBootstrap(h host.Host, bootstrapPeers []string, minPeers int) (error, Bootstrap) {

	if minPeers > len(bootstrapPeers) {
		return errors.New(fmt.Sprintf("Too less bootstrapping nodes. Expected at least: %d, got: %d", minPeers, len(bootstrapPeers))), Bootstrap{}
	}

	return nil, Bootstrap{
		minPeers:       minPeers,
		bootstrapPeers: bootstrapPeers,
		host:           h,
	}

}
