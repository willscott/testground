package test

import (
	"context"
	"time"

	coreopts "github.com/ipfs/interface-go-ipfs-core/options"
	utils "github.com/ipfs/testground/plans/chew-datasets/utils"
	"github.com/ipfs/testground/sdk/iptb"
	"github.com/ipfs/testground/sdk/runtime"
)

// IpfsAddTrickleDag IPFS Add Trickle DAG Test
type IpfsAddTrickleDag struct{}

func (t *IpfsAddTrickleDag) AcceptFiles() bool {
	return true
}

func (t *IpfsAddTrickleDag) AcceptDirs() bool {
	return false
}

func (t *IpfsAddTrickleDag) AddRepoOptions() iptb.AddRepoOptions {
	return nil
}

func (t *IpfsAddTrickleDag) Execute(ctx context.Context, runenv *runtime.RunEnv, cfg *utils.TestCaseOptions) {
	err := cfg.ForEachPath(runenv, func(path string, size int64, isDir bool) (string, error) {
		unixfsFile, err := utils.ConvertToUnixfs(path, isDir)
		if err != nil {
			return "", err
		}

		addOptions := func(settings *coreopts.UnixfsAddSettings) error {
			settings.Layout = coreopts.TrickleLayout
			return nil
		}

		tstarted := time.Now()
		cidFile, err := cfg.API.Unixfs().Add(ctx, unixfsFile, addOptions)
		if err != nil {
			return "", err
		}
		runenv.EmitMetric(utils.MakeTimeToAddMetric(size, cfg.Mode), float64(time.Now().Sub(tstarted)/time.Millisecond))

		return cidFile.String(), nil
	})

	if err != nil {
		runenv.Abort(err)
		return
	}

	runenv.OK()
}
