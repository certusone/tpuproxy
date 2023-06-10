package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"go.firedancer.io/radiance/cmd/radiance/tpu_quic"
	"go.firedancer.io/radiance/cmd/radiance/tpu_udp"

	"github.com/spf13/cobra"
	"go.firedancer.io/radiance/cmd/radiance/blockstore"
	"go.firedancer.io/radiance/cmd/radiance/gossip"
	"go.firedancer.io/radiance/cmd/radiance/replay"
	"k8s.io/klog/v2"

	// Load in instruction pretty-printing
	_ "github.com/gagliardetto/solana-go/programs/system"
	_ "github.com/gagliardetto/solana-go/programs/vote"
)

var cmd = cobra.Command{
	Use:   "radiance",
	Short: "Solana Go playground",
}

func init() {
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	cmd.PersistentFlags().AddGoFlagSet(klogFlags)

	cmd.AddCommand(
		&blockstore.Cmd,
		&gossip.Cmd,
		&replay.Cmd,
		&tpu_udp.Cmd,
		&tpu_quic.Cmd,
	)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	cobra.CheckErr(cmd.ExecuteContext(ctx))
}
