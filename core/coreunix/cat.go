package coreunix

import (
	"context"

	core "github.com/ipfs/go-ipfs/core"
	uio "gx/ipfs/QmY8ZHzFG4HVQqUG1L7MDrSuZUcRWJviuU6YLCuy8BLEcL/go-unixfs/io"
	path "gx/ipfs/QmcQtKwefUJDALNf2CSPw4CAfyjyhwFY5yWnnWuGxFdJCc/go-path"
	resolver "gx/ipfs/QmcQtKwefUJDALNf2CSPw4CAfyjyhwFY5yWnnWuGxFdJCc/go-path/resolver"
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
