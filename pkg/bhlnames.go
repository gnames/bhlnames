package bhlnames

import (
	"fmt"

	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/ent/reffnd"
	"github.com/gnames/bhlnames/internal/ent/ttlmch"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
)

// Option provides an 'interface' for setting up BHLnames instance.
type Option func(*bhlnames)

// OptBuilder sets the Builder of initialization process.
func OptBuilder(b builder.Builder) Option {
	return func(bn *bhlnames) {
		bn.bld = b
	}
}

// bhlnames implements BHLnames interface.
type bhlnames struct {
	// cfg is a configuration for BHLnames.
	cfg config.Config

	// bld is a Builder for BHLnames. Builder is used only for the
	// initialization process.
	bld builder.Builder

	// rf is a RefFinder for BHLnames. RefFinder is used for finding
	// references in BHL according to input.
	rf reffnd.RefFinder

	// tm is a TitleMatcher for BHLnames. TitleMatcher is used for
	// finding possible matches to a reference title from the input.
	tm ttlmch.TitleMatcher

	gnpPool chan gnparser.GNparser
}

// New creates a new BHLnames instance.
func New(cfg config.Config, opts ...Option) BHLnames {
	res := bhlnames{cfg: cfg}
	for _, opt := range opts {
		opt(&res)
	}

	res.gnPool = gnparserPool(cfg.JobsNum)
	return &res
}

// Initialize downloads BHL's metadata and imports it into the storage.
func (bn bhlnames) Initialize() error {
	var err error
	if bn.cfg.WithRebuild {
		bn.bld.ResetData()
	}

	err = bn.bld.ImportData()
	if err != nil {
		err = fmt.Errorf("ImportData: %w", err)
		return err
	}

	err = bn.bld.CalculateTxStats()
	if err != nil {
		err = fmt.Errorf("CalculateTxStats: %w", err)
		return err
	}

	return bn.Close()
}

func (bn *bhlnames) Close() error {
	if bn.bld != nil {
		bn.bld.Close()
	}
	return nil
}

func gnparserPool(poolSize int) chan gnparser.GNparser {
	gnpPool := make(chan gnparser.GNparser, poolSize)
	for i := 0; i < poolSize; i++ {
		cfgGNP := gnparser.NewConfig()
		gnpPool <- gnparser.New(cfgGNP)
	}
	return gnpPool
}
