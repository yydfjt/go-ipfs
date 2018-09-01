package coreapi_test

import (
	"context"
	"io"
	"testing"

	"github.com/ipfs/go-ipfs/core/coreapi/interface/options"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
)

func TestDhtFindPeer(t *testing.T) {
	ctx := context.Background()
	nds, apis, err := makeAPISwarm(ctx, true, 5)
	if err != nil {
		t.Fatal(err)
	}

	pi, err := apis[2].Dht().FindPeer(ctx, peer.ID(nds[0].Identity))
	if err != nil {
		t.Fatal(err)
	}

	if pi.Addrs[0].String() != "/ip4/127.0.0.1/tcp/4001" {
		t.Errorf("got unexpected address from FindPeer: %s", pi.Addrs[0].String())
	}

	pi, err = apis[1].Dht().FindPeer(ctx, peer.ID(nds[2].Identity))
	if err != nil {
		t.Fatal(err)
	}

	if pi.Addrs[0].String() != "/ip4/127.0.2.1/tcp/4001" {
		t.Errorf("got unexpected address from FindPeer: %s", pi.Addrs[0].String())
	}
}

func TestDhtFindProviders(t *testing.T) {
	ctx := context.Background()
	nds, apis, err := makeAPISwarm(ctx, true, 5)
	if err != nil {
		t.Fatal(err)
	}

	p, err := addTestObject(ctx, apis[0])
	if err != nil {
		t.Fatal(err)
	}

	out, err := apis[2].Dht().FindProviders(ctx, p, options.Dht.NumProviders(1))
	if err != nil {
		t.Fatal(err)
	}

	provider := <-out

	if provider.ID.String() != nds[0].Identity.String() {
		t.Errorf("got wrong provider: %s != %s", provider.ID.String(), nds[0].Identity.String())
	}
}

func TestDhtProvide(t *testing.T) {
	ctx := context.Background()
	nds, apis, err := makeAPISwarm(ctx, true, 5)
	if err != nil {
		t.Fatal(err)
	}

	p1, err := apis[0].Block().Put(ctx, &io.LimitedReader{R: rnd, N: 4092}, options.Block.Format("raw"))
	p2, err := apis[0].Block().Put(ctx, &io.LimitedReader{R: rnd, N: 4092}, options.Block.Format("raw"))

	out, err := apis[2].Dht().FindProviders(ctx, p1, options.Dht.NumProviders(1))
	if err != nil {
		t.Fatal(err)
	}

	provider := <-out

	if provider.ID.String() != "<peer.ID >" {
		t.Errorf("got wrong provider: %s != %s", provider.ID.String(), nds[0].Identity.String())
	}

	err = apis[0].Dht().Provide(ctx, p1)
	if err != nil {
		t.Fatal(err)
	}

	out, err = apis[2].Dht().FindProviders(ctx, p1, options.Dht.NumProviders(1))
	if err != nil {
		t.Fatal(err)
	}

	provider = <-out

	if provider.ID.String() != nds[0].Identity.String() {
		t.Errorf("got wrong provider: %s != %s", provider.ID.String(), nds[0].Identity.String())
	}

	err = apis[0].Dht().Provide(ctx, p2, options.Dht.Recursive(true))
	if err != nil {
		t.Fatal(err)
	}

	out, err = apis[2].Dht().FindProviders(ctx, p2, options.Dht.NumProviders(1))
	if err != nil {
		t.Fatal(err)
	}

	provider = <-out

	if provider.ID.String() != nds[0].Identity.String() {
		t.Errorf("got wrong provider: %s != %s", provider.ID.String(), nds[0].Identity.String())
	}
}
