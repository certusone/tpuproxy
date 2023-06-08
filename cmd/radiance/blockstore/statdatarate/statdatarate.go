//go:build !lite

package statdatarate

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/klauspost/compress/zstd"
	"github.com/spf13/cobra"
	"go.firedancer.io/radiance/cmd/radiance/blockstore/util"
	"go.firedancer.io/radiance/pkg/blockstore"
	"k8s.io/klog/v2"
)

var Cmd = cobra.Command{
	Use:   "stat-data-rate <rocksdb> <slots>",
	Short: "Produce CSV report of data rate at slot-level granularity",
	Args:  cobra.ExactArgs(2),
}

func init() {
	Cmd.Run = run
}

func run(_ *cobra.Command, args []string) {
	slots, ok := util.ParseInts(args[1])
	if !ok {
		klog.Exit("Invalid slots parameter")
	}

	db, err := blockstore.OpenReadOnly(args[0])
	if err != nil {
		klog.Exitf("Failed to open blockstore: %s", err)
	}
	defer db.Close()

	wr := csv.NewWriter(os.Stdout)
	defer wr.Flush()
	wr.Write([]string{"slot", "ts", "block_raw_bytes", "block_compressed_bytes"})

	slots.Iter(func(slot uint64) bool {
		err := dumpSlot(db, wr, slot)
		if err != nil {
			klog.Warningf("Failed to dump slot %d: %s", slot, err)
		}
		return true
	})
}

func dumpSlot(db *blockstore.DB, wr *csv.Writer, slot uint64) error {
	slotDecimal := strconv.FormatUint(slot, 10)

	meta, err := db.GetSlotMeta(slot)
	if err != nil {
		return err
	}
	entries, err := db.GetEntries(meta, 2)
	if err != nil {
		return err
	}

	var compressedSize countWriter
	compressor, err := zstd.NewWriter(&compressedSize)
	if err != nil {
		panic(err.Error())
	}
	var blockRawBytes uint64
	for _, batch := range entries {
		blockRawBytes += uint64(len(batch.Raw))
		_, err := compressor.Write(batch.Raw)
		if err != nil {
			return err
		}
	}
	_ = compressor.Flush()

	wr.Write([]string{
		slotDecimal,
		strconv.FormatUint(meta.FirstShredTimestamp, 10),
		strconv.FormatUint(uint64(blockRawBytes), 10),
		strconv.FormatUint(compressedSize.n, 10),
	})

	return nil
}

type countWriter struct {
	n uint64
}

func (c *countWriter) Write(p []byte) (int, error) {
	c.n += uint64(len(p))
	return len(p), nil
}
