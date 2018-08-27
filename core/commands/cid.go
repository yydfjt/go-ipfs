package commands

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode"

	cmds "gx/ipfs/QmPTfgFTo9PFr1PvPKyKoeMgBvYPh6cX3aDP7DHKVbnCbi/go-ipfs-cmds"
	mhash "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
	cidutil "gx/ipfs/QmPyxJ2QS7L5FhGkNYkNcXHGjDhvGHueJ4auqAstFHYxy5/go-cidutil"
	cmdkit "gx/ipfs/QmSP88ryZkHSRn1fnngAaV2Vcn63WUJzAavnRM9CVdU1Ky/go-ipfs-cmdkit"
	mbase "gx/ipfs/QmSbvata2WqNkqGtZNg8MR3SKwnB8iQ7vTPJgWqB8bC5kR/go-multibase"
	verifcid "gx/ipfs/QmVUhfewLZpSaAiBYCpw2krYMaiVmFuhr2iurQLuRoU6sD/go-verifcid"
	cid "gx/ipfs/QmZFbDTY9jfSBms2MchvYM9oYRbAF19K7Pby47yDBfpPrb/go-cid"
)

var CidCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Convert and discover properties of CIDs",
	},
	Subcommands: map[string]*cmds.Command{
		"format": cidFmtCmd,
		"base32": base32Cmd,
		"bases":  basesCmd,
		"codecs": codecsCmd,
		"hashes": hashesCmd,
	},
}

var cidFmtCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Format and convert a CID in various useful ways.",
		LongDescription: `
Format and converts <cid>'s in various useful ways.

The optional format string is a printf style format string:
` + cidutil.FormatRef,
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("cid", true, true, "Cids to format."),
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption("f", "Printf style format string.").WithDefault("%s"),
		cmdkit.StringOption("v", "CID version to convert to."),
		cmdkit.StringOption("b", "Multibase to display CID in."),
	},
	Run: func(req *cmds.Request, resp cmds.ResponseEmitter, env cmds.Environment) {
		fmtStr, _ := req.Options["f"].(string)
		verStr, _ := req.Options["v"].(string)
		baseStr, _ := req.Options["b"].(string)

		opts := cidFormatOpts{}

		if strings.IndexByte(fmtStr, '%') == -1 {
			resp.SetError(fmt.Errorf("invalid format string: %s", fmtStr), cmdkit.ErrNormal)
			return
		}
		opts.fmtStr = fmtStr

		switch verStr {
		case "":
			// noop
		case "0":
			opts.verConv = toCidV0
		case "1":
			opts.verConv = toCidV1
		default:
			resp.SetError(fmt.Errorf("invalid cid version: %s\n", verStr), cmdkit.ErrNormal)
			return
		}

		if baseStr != "" {
			encoder, err := mbase.EncoderByName(baseStr)
			if err != nil {
				resp.SetError(err, cmdkit.ErrNormal)
				return
			}
			opts.newBase = encoder.Encoding()
		} else {
			opts.newBase = mbase.Encoding(-1)
		}

		emitCids(req, resp, opts)
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: streamRes(func(v interface{}, out io.Writer) nonFatalError {
			r := v.(*CidFormatRes)
			if r.ErrorMsg != "" {
				return nonFatalError(fmt.Sprintf("%s: %s", r.CidStr, r.ErrorMsg))
			}
			fmt.Fprintf(out, "%s\n", r.Formatted)
			return ""
		}),
	},
	Type: CidFormatRes{},
}

type CidFormatRes struct {
	CidStr    string // Original Cid String passed in
	Formatted string // Formated Result
	ErrorMsg  string // Error
}

var base32Cmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Convert CIDs to Base32 CID version 1.",
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("cid", true, true, "Cids to convert.").EnableStdin(),
	},
	Run: func(req *cmds.Request, resp cmds.ResponseEmitter, env cmds.Environment) {
		opts := cidFormatOpts{
			fmtStr:  "%s",
			newBase: mbase.Encoding(mbase.Base32),
			verConv: toCidV1,
		}
		emitCids(req, resp, opts)
	},
	PostRun: cidFmtCmd.PostRun,
	Type:    cidFmtCmd.Type,
}

type cidFormatOpts struct {
	fmtStr  string
	newBase mbase.Encoding
	verConv func(cid *cid.Cid) (*cid.Cid, error)
}

func emitCids(req *cmds.Request, resp cmds.ResponseEmitter, opts cidFormatOpts) {
	for _, cidStr := range req.Arguments {
		emit := func(fmtd string, err error) {
			res := &CidFormatRes{CidStr: cidStr, Formatted: fmtd}
			if err != nil {
				res.ErrorMsg = err.Error()
			}
			resp.Emit(res)
		}
		c, err := cid.Decode(cidStr)
		if err != nil {
			emit("", err)
			continue
		}
		base := opts.newBase
		if base == -1 {
			base, _ = cid.ExtractEncoding(cidStr)
		}
		if opts.verConv != nil {
			c, err = opts.verConv(c)
			if err != nil {
				emit("", err)
				continue
			}
		}
		str, err := cidutil.Format(opts.fmtStr, base, c)
		if _, ok := err.(cidutil.FormatStringError); ok {
			// no point in continuing if there is a problem with the format string
			resp.SetError(err, cmdkit.ErrNormal)
			return
		}
		emit(str, err)
	}
}

func toCidV0(c *cid.Cid) (*cid.Cid, error) {
	if c.Type() != cid.DagProtobuf {
		return nil, fmt.Errorf("can't convert non-protobuf nodes to cidv0")
	}
	return cid.NewCidV0(c.Hash()), nil
}

func toCidV1(c *cid.Cid) (*cid.Cid, error) {
	return cid.NewCidV1(c.Type(), c.Hash()), nil
}

type CodeAndName struct {
	Code int
	Name string
}

var basesCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "List available multibase encodings.",
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("prefix", "also include the single leter prefixes in addition to the code"),
		cmdkit.BoolOption("numeric", "also include numeric codes"),
	},
	Run: func(req *cmds.Request, resp cmds.ResponseEmitter, env cmds.Environment) {
		var res []CodeAndName
		// use EncodingToStr in case at some point there are multiple names for a given code
		for code, name := range mbase.EncodingToStr {
			res = append(res, CodeAndName{int(code), name})
		}
		cmds.EmitOnce(resp, res)
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w0 io.Writer, val0 interface{}) error {
			w := tabwriter.NewWriter(w0, 0, 0, 2, ' ', 0)
			prefixes, _ := req.Options["prefix"].(bool)
			numeric, _ := req.Options["numeric"].(bool)
			val := val0.([]CodeAndName)
			sort.Sort(multibaseSorter{val})
			for _, v := range val {
				if prefixes && v.Code >= 32 && v.Code < 127 {
					fmt.Fprintf(w, "%c\t", v.Code)
				} else if prefixes {
					// don't display non-printable prefixes
					fmt.Fprintf(w, "\t")
				}
				if numeric {
					fmt.Fprintf(w, "%d\t%s\n", v.Code, v.Name)
				} else {
					fmt.Fprintf(w, "%s\n", v.Name)
				}
			}
			w.Flush()
			return nil
		}),
	},
	Type: []CodeAndName{},
}

var codecsCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "List available CID codecs.",
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("numeric", "also include numeric codes"),
	},
	Run: func(req *cmds.Request, resp cmds.ResponseEmitter, env cmds.Environment) {
		var res []CodeAndName
		// use CodecToStr as there are multiple names for a given code
		for code, name := range cid.CodecToStr {
			res = append(res, CodeAndName{int(code), name})
		}
		cmds.EmitOnce(resp, res)
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w0 io.Writer, val0 interface{}) error {
			w := tabwriter.NewWriter(w0, 0, 0, 2, ' ', 0)
			numeric, _ := req.Options["numeric"].(bool)
			val := val0.([]CodeAndName)
			sort.Sort(codeAndNameSorter{val})
			for _, v := range val {
				if numeric {
					fmt.Fprintf(w, "%d\t%s\n", v.Code, v.Name)
				} else {
					fmt.Fprintf(w, "%s\n", v.Name)
				}
			}
			w.Flush()
			return nil
		}),
	},
	Type: []CodeAndName{},
}

var hashesCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "List available multihashes.",
	},
	Options: codecsCmd.Options,
	Run: func(req *cmds.Request, resp cmds.ResponseEmitter, env cmds.Environment) {
		var res []CodeAndName
		// use mhash.Codes in case at some point there are multiple names for a given code
		for code, name := range mhash.Codes {
			if !verifcid.IsGoodHash(code) {
				continue
			}
			res = append(res, CodeAndName{int(code), name})
		}
		cmds.EmitOnce(resp, res)
	},
	Encoders: codecsCmd.Encoders,
	Type:     codecsCmd.Type,
}

type multibaseSorter struct {
	data []CodeAndName
}

func (s multibaseSorter) Len() int      { return len(s.data) }
func (s multibaseSorter) Swap(i, j int) { s.data[i], s.data[j] = s.data[j], s.data[i] }

func (s multibaseSorter) Less(i, j int) bool {
	a := unicode.ToLower(rune(s.data[i].Code))
	b := unicode.ToLower(rune(s.data[j].Code))
	if a != b {
		return a < b
	}
	// lowecase letters should come before uppercase
	return s.data[i].Code > s.data[j].Code
}

type codeAndNameSorter struct {
	data []CodeAndName
}

func (s codeAndNameSorter) Len() int           { return len(s.data) }
func (s codeAndNameSorter) Swap(i, j int)      { s.data[i], s.data[j] = s.data[j], s.data[i] }
func (s codeAndNameSorter) Less(i, j int) bool { return s.data[i].Code < s.data[j].Code }
