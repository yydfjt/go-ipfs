package coreunix

import (
	"context"

	core "github.com/ipfs/go-ipfs/core"
	path "gx/ipfs/QmSx7Fv8e2QenkYqRP865pTaMEMpwjmnyZqJXTfAwRuiBU/go-path"
	resolver "gx/ipfs/QmSx7Fv8e2QenkYqRP865pTaMEMpwjmnyZqJXTfAwRuiBU/go-path/resolver"
	uio "gx/ipfs/QmeoBC7eiuWuMvRwYNYg5rBHZk1rizyfnsMBrkojhrPNkX/go-unixfs/io"
)

func Cat(ctx context.Context, n *core.IpfsNode, pstr string) (uio.DagReader, error) {
	r := &resolver.Resolver{
		DAG:         n.DAG,
		ResolveOnce: uio.ResolveUnixfsOnce,
	}

	dagNode, err := core.Resolve(ctx, n.Namesys, r, path.Path(pstr))
	if err != nil {
		return nil, err
	}

	return uio.NewDagReader(ctx, dagNode, n.DAG)
}
