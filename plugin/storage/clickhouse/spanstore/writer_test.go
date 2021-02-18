package spanstore_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/pkg/config"
	"github.com/jaegertracing/jaeger/plugin/storage/clickhouse"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

func BenchmarkWrites(b *testing.B) {
	runFactoryTest(b, func(tb testing.TB, mock sqlmock.Sqlmock, sw spanstore.Writer, sr spanstore.Reader) {
		spans := 32
		tagsCount := 64
		tags, services, operations := makeWriteSupports(tagsCount, spans)

		s := model.Span{
			TraceID: model.TraceID{
				Low:  rand.Uint64(),
				High: 0,
			},
			SpanID:        model.SpanID(rand.Int()),
			OperationName: operations[rand.Intn(spans)],
			Process: &model.Process{
				ServiceName: services[rand.Intn(spans)],
			},
			Tags:      tags,
			StartTime: time.Now().Add(time.Millisecond),
			Duration:  time.Millisecond,
		}

		b.ResetTimer()
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = sw.WriteSpan(context.Background(), &s)
			}
		})
		b.StopTimer()
	})
}

func makeWriteSupports(tagsCount, spans int) ([]model.KeyValue, []string, []string) {
	tags := make([]model.KeyValue, tagsCount)
	for i := 0; i < tagsCount; i++ {
		tags[i] = model.KeyValue{
			Key:  fmt.Sprintf("a%d", i),
			VStr: fmt.Sprintf("b%d", i),
		}
	}
	operations := make([]string, spans)
	for j := 0; j < spans; j++ {
		operations[j] = fmt.Sprintf("operation-%d", j)
	}
	services := make([]string, spans)
	for i := 0; i < spans; i++ {
		services[i] = fmt.Sprintf("service-%d", i)
	}

	return tags, services, operations
}

// Benchmarks intended for profiling
func writeSpans(sw spanstore.Writer, tags []model.KeyValue, services, operations []string, traces, spans int, high uint64, tid time.Time) {
	for i := 0; i < traces; i++ {
		for j := 0; j < spans; j++ {
			s := model.Span{
				TraceID: model.TraceID{
					Low:  uint64(i),
					High: high,
				},
				SpanID:        model.SpanID(j),
				OperationName: operations[j],
				Process: &model.Process{
					ServiceName: services[j],
				},
				Tags:      tags,
				StartTime: tid.Add(time.Millisecond),
				Duration:  time.Millisecond * time.Duration(i+j),
			}
			_ = sw.WriteSpan(context.Background(), &s)
		}
	}
}

func fakeConnector() (sqlmock.Sqlmock, clickhouse.Connector, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	connector := func(config *clickhouse.NamespaceConfig) (*sql.DB, error) {
		return db, err
	}

	return mock, connector, err
}

// Fake a sql.DB and runs a test on it.
func runFactoryTest(tb testing.TB, test func(tb testing.TB, mock sqlmock.Sqlmock, sw spanstore.Writer, sr spanstore.Reader)) {
	mock, connector, err := fakeConnector()
	assert.NoError(tb, err)

	clickhouse.WithConnector(connector)

	f := clickhouse.NewFactory()
	//defer func() {
	//	require.NoError(tb, f.Close())
	//}()

	opts := clickhouse.NewOptions("clickhouse")
	v, command := config.Viperize(opts.AddFlags)
	command.ParseFlags([]string{
		"--badger.ephemeral=true",
		"--badger.consistency=false",
	})
	f.InitFromViper(v)

	err = f.Initialize(metrics.NullFactory, zap.NewNop())
	assert.NoError(tb, err)

	sw, err := f.CreateSpanWriter()
	assert.NoError(tb, err)

	sr, err := f.CreateSpanReader()
	assert.NoError(tb, err)

	test(tb, mock, sw, sr)
}
