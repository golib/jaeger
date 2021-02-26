package clickhouse

import (
	"flag"
	"time"

	"github.com/spf13/viper"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/jaegertracing/jaeger/plugin/storage/clickhouse/spanstore"
)

const (
	defaultDatasource        string             = "tcp://localhost:9000"
	defaultOperationsTable   string             = "jaeger_operations_v2"
	defaultIndexTable        string             = "jaeger_index_v2"
	defaultSpansTable        string             = "jaeger_spans_v2"
	defaultArchiveSpansTable string             = "jaeger_archive_spans_v2"
	defaultWriteBatchDelay   time.Duration      = 5 * time.Second
	defaultWriteBatchSize    int                = 10000
	defaultEncoding          spanstore.Encoding = spanstore.EncodingProto
)

const (
	suffixEnabled         = ".enabled"
	suffixDatasource      = ".datasource"
	suffixOperationsTable = ".operations-table"
	suffixIndexTable      = ".index-table"
	suffixSpansTable      = ".spans-table"
	suffixWriteBatchDelay = ".write-batch-delay"
	suffixWriteBatchSize  = ".write-batch-size"
	suffixEncoding        = ".encoding"
)

// NamespaceConfig is Clickhouse's internal configuration data
type NamespaceConfig struct {
	namespace       string
	Enabled         bool
	Datasource      string
	OperationsTable string
	IndexTable      string
	SpansTable      string
	WriteBatchDelay time.Duration
	WriteBatchSize  int
	Encoding        spanstore.Encoding
	Connector       Connector
}

// Options store storage plugin related configs
type Options struct {
	primary *NamespaceConfig

	others map[string]*NamespaceConfig
}

// NewOptions creates a new Options struct.
func NewOptions(primaryNamespace string, otherNamespaces ...string) *Options {
	options := &Options{
		primary: &NamespaceConfig{
			namespace:       primaryNamespace,
			Enabled:         true,
			Datasource:      defaultDatasource,
			OperationsTable: defaultOperationsTable,
			IndexTable:      defaultIndexTable,
			SpansTable:      defaultSpansTable,
			WriteBatchDelay: defaultWriteBatchDelay,
			WriteBatchSize:  defaultWriteBatchSize,
			Encoding:        defaultEncoding,
			Connector:       defaultConnector,
		},
		others: make(map[string]*NamespaceConfig, len(otherNamespaces)),
	}

	for _, namespace := range otherNamespaces {
		if namespace == archiveNamespace {
			options.others[namespace] = &NamespaceConfig{
				namespace:       namespace,
				Datasource:      defaultDatasource,
				OperationsTable: "",
				IndexTable:      "",
				SpansTable:      defaultArchiveSpansTable,
				WriteBatchDelay: defaultWriteBatchDelay,
				WriteBatchSize:  defaultWriteBatchSize,
				Encoding:        defaultEncoding,
				Connector:       defaultConnector,
			}
		} else {
			options.others[namespace] = &NamespaceConfig{namespace: namespace}
		}
	}

	return options
}

// AddFlags adds flags for Options
func (opt *Options) AddFlags(flagSet *flag.FlagSet) {
	addFlags(flagSet, opt.primary)
	for _, cfg := range opt.others {
		addFlags(flagSet, cfg)
	}
}

func addFlags(flagSet *flag.FlagSet, nsConfig *NamespaceConfig) {
	if nsConfig.namespace == archiveNamespace {
		flagSet.Bool(
			nsConfig.namespace+suffixEnabled,
			nsConfig.Enabled,
			"Enable archive storage")
	}

	flagSet.String(
		nsConfig.namespace+suffixDatasource,
		nsConfig.Datasource,
		"Clickhouse datasource string.",
	)

	if nsConfig.namespace != archiveNamespace {
		flagSet.String(
			nsConfig.namespace+suffixOperationsTable,
			nsConfig.OperationsTable,
			"Clickhouse operations table name.",
		)

		flagSet.String(
			nsConfig.namespace+suffixIndexTable,
			nsConfig.IndexTable,
			"Clickhouse index table name.",
		)
	}

	flagSet.String(
		nsConfig.namespace+suffixSpansTable,
		nsConfig.SpansTable,
		"Clickhouse spans table name.",
	)

	flagSet.Duration(
		nsConfig.namespace+suffixWriteBatchDelay,
		nsConfig.WriteBatchDelay,
		"A duration after which spans are flushed to Clickhouse",
	)

	flagSet.Int(
		nsConfig.namespace+suffixWriteBatchSize,
		nsConfig.WriteBatchSize,
		"A number of spans buffered before they are flushed to Clickhouse",
	)

	flagSet.String(
		nsConfig.namespace+suffixEncoding,
		string(nsConfig.Encoding),
		"Encoding to store spans (json allows out of band queries, protobuf is more compact)",
	)
}

// InitFromViper initializes Options with properties from viper
func (opt *Options) InitFromViper(v *viper.Viper) {
	initFromViper(opt.primary, v)
	for _, cfg := range opt.others {
		initFromViper(cfg, v)
	}
}

func initFromViper(cfg *NamespaceConfig, v *viper.Viper) {
	cfg.Enabled = v.GetBool(cfg.namespace + suffixEnabled)
	cfg.Datasource = v.GetString(cfg.namespace + suffixDatasource)
	cfg.IndexTable = v.GetString(cfg.namespace + suffixIndexTable)
	cfg.SpansTable = v.GetString(cfg.namespace + suffixSpansTable)
	cfg.OperationsTable = v.GetString(cfg.namespace + suffixOperationsTable)
	cfg.WriteBatchDelay = v.GetDuration(cfg.namespace + suffixWriteBatchDelay)
	cfg.WriteBatchSize = v.GetInt(cfg.namespace + suffixWriteBatchSize)
	cfg.Encoding = spanstore.Encoding(v.GetString(cfg.namespace + suffixEncoding))
}

// GetPrimary returns the primary namespace configuration
func (opt *Options) getPrimary() *NamespaceConfig {
	return opt.primary
}
