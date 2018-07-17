package commands

import (
	core "github.com/ipfs/go-ipfs/core"
	path "github.com/ipfs/go-ipfs/path"
	resolver "github.com/ipfs/go-ipfs/path/resolver"
	uio "github.com/ipfs/go-ipfs/unixfs/io"

	cmds "gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	"gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit"
)

var FastLsCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "and ls that waits for no one",
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("ipfs-path", true, true, "The path to the IPFS object(s) to list links from.").EnableStdin(),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) {
		nd, err := GetNode(env)
		if err != nil {
			re.SetError(err, cmdkit.ErrNormal)
			return
		}

		var dagnodes []ipld.Node
		for _, fpath := range req.Arguments {
			p, err := path.ParsePath(fpath)
			if err != nil {
				re.SetError(err, cmdkit.ErrNormal)
				return
			}

			r := &resolver.Resolver{
				DAG:         nd.DAG,
				ResolveOnce: uio.ResolveUnixfsOnce,
			}

			dagnode, err := core.Resolve(req.Context, nd.Namesys, r, p)
			if err != nil {
				re.SetError(err, cmdkit.ErrNormal)
				return
			}
			dagnodes = append(dagnodes, dagnode)
		}

		for _, dagnode := range dagnodes {
			dir, err := uio.NewDirectoryFromNode(nd.DAG, dagnode)
			if err != nil && err != uio.ErrNotADir {
				re.SetError(err, cmdkit.ErrNormal)
				return
			}

			err = dir.ForEachLink(req.Context, func(l *ipld.Link) error {
				re.Emit(l.Name)
				return nil
			})
			if err != nil {
				re.SetError(err, cmdkit.ErrNormal)
				return
			}
		}
	},
	Encoders: cmds.EncoderMap{},
	Type:     string(""),
}
