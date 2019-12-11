package iptb

import (
	"context"

	httpapi "github.com/ipfs/go-ipfs-http-client"
	iface "github.com/ipfs/interface-go-ipfs-core"
	ma "github.com/multiformats/go-multiaddr"
)

// SpawnDaemon spawns a daemon using the InterPlanetary Test Bed and returns
// the ensemble (you must call ensemble.Destroy() in the end) and the client API
// connection.
func SpawnDaemon(ctx context.Context, opts NodeOpts) (*TestEnsemble, iface.CoreAPI) {
	spec := NewTestEnsembleSpec()
	spec.AddNodesDefaultConfig(opts, "node")

	ensemble := NewTestEnsemble(ctx, spec)
	ensemble.Initialize()

	node := ensemble.GetNode("node")

	addr, err := node.APIAddr()
	if err != nil {
		panic(err)
	}

	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		panic(err)
	}

	api, err := httpapi.NewApi(maddr)
	if err != nil {
		panic(err)
	}

	return ensemble, api
}
