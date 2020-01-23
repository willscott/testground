package test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/ipfs/testground/sdk/runtime"
	"github.com/ipfs/testground/sdk/sync"
)

func iproute() {
	cmd := exec.Command("ip", "route")
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(string(stdout))
}

func FindPeers(runenv *runtime.RunEnv) {
	runenv.Message("starting findPeers")
	opts := &SetupOpts{
		Timeout:     time.Duration(runenv.IntParam("timeout_secs")) * time.Second,
		RandomWalk:  runenv.BooleanParam("random_walk"),
		NBootstrap:  runenv.IntParam("n_bootstrap"),
		NFindPeers:  runenv.IntParam("n_find_peers"),
		BucketSize:  runenv.IntParam("bucket_size"),
		AutoRefresh: runenv.BooleanParam("auto_refresh"),
	}

	if opts.NFindPeers > runenv.TestInstanceCount {
		runenv.Abort("NFindPeers greater than the number of test instances")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	//runenv.Message("ip route - first")
	//iproute()
	//time.Sleep(5 * time.Second)
	//runenv.Message("ip route - second")
	//iproute()

	runenv.Message("before MustWatcherWriter")
	watcher, writer := sync.MustWatcherWriter(runenv)
	defer watcher.Close()
	defer writer.Close()

	runenv.Message("before Setup")
	_, dht, peers, seq, err := Setup(ctx, runenv, watcher, writer, opts)
	if err != nil {
		runenv.Abort(err)
		return
	}

	//time.Sleep(5 * time.Second)
	//runenv.Message("ip route - after setup")
	//iproute()

	defer Teardown(ctx, runenv, watcher, writer)

	runenv.Message("before Bootstrap")
	// Bring the network into a nice, stable, bootstrapped state.
	if err = Bootstrap(ctx, runenv, watcher, writer, opts, dht, peers, seq); err != nil {
		runenv.Abort(err)
		return
	}

	if opts.RandomWalk {
		if err = RandomWalk(ctx, runenv, dht); err != nil {
			runenv.Abort(err)
			return
		}
	}

	// Ok, we're _finally_ ready.
	// TODO: Dump routing table stats. We should dump:
	//
	// * How full our "closest" bucket is. That is, look at the "all peers"
	//   list, find the BucketSize closest peers, and determine the % of those
	//   peers to which we're connected. It should be close to 100%.
	// * How many peers we're actually connected to?
	// * How many of our connected peers are actually useful to us?

	// Perform FIND_PEER N times.
	found := 0
	for _, p := range peers {
		if found >= opts.NFindPeers {
			break
		}
		if len(dht.Host().Peerstore().Addrs(p.ID)) > 0 {
			// Skip peer's we've already found (even if we've
			// disconnected for some reason).
			continue
		}

		t := time.Now()

		// TODO: Instrument libp2p dht to get:
		// - Number of peers dialed
		// - Number of dials along the way that failed
		if _, err := dht.FindPeer(ctx, p.ID); err != nil {
			runenv.Abort(fmt.Errorf("find peer failed: %s", err))
			return
		}

		runenv.EmitMetric(&runtime.MetricDefinition{
			Name:           fmt.Sprintf("time-to-find-%d", found),
			Unit:           "ns",
			ImprovementDir: -1,
		}, float64(time.Now().Sub(t).Nanoseconds()))

		found++
	}
	runenv.OK()
}
