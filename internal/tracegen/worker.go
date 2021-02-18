// Copyright (c) 2018 The Jaeger Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracegen

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"

	jaegerConfig "github.com/uber/jaeger-client-go/config"
	jaegerZap "github.com/uber/jaeger-client-go/log/zap"

	"github.com/icrowley/fake"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
)

type worker struct {
	running         *uint32         // pointer to shared flag that indicates it's time to stop the test
	serviceApis     []string        // faked api list of services
	chainedServices []string        // nested services
	id              int             // worker id
	traces          int             // how many traces the worker has to generate (only when duration==0)
	marshal         bool            // whether the worker needs to marshal trace context via HTTP headers
	debug           bool            // whether to set DEBUG flag on the spans
	firehose        bool            // whether to set FIREHOSE flag on the spans
	duration        time.Duration   // how long to run the test for (overrides `traces`)
	pause           time.Duration   // how long to pause before finishing the trace
	wg              *sync.WaitGroup // notify when done
	logger          *zap.Logger
}

var (
	fakeIP = func() uint32 {
		ipv4 := net.ParseIP(fake.IPv4()).To4()

		return uint32(ipv4[3]) | uint32(ipv4[2])<<8 | uint32(ipv4[1])<<16 | uint32(ipv4[0])<<24
	}

	fakeSpanDuration = func() time.Duration {
		return time.Duration(rand.Intn(10)+rand.Intn(100)) * time.Millisecond
	}

	fakePausedDuration = func(spanDuration time.Duration) time.Duration {
		pause := time.Duration(rand.Intn(3+int(spanDuration/time.Millisecond))) * time.Millisecond
		time.Sleep(pause)

		return pause
	}

	fakeTracerStore  = sync.Map{}
	fakeTracerSingle = new(singleflight.Group)

	fakeTracer = func(service string) opentracing.Tracer {
		value, _, _ := fakeTracerSingle.Do(service, func() (interface{}, error) {
			value, ok := fakeTracerStore.Load(service)
			if ok {
				return value, nil
			}

			var logger, _ = zap.NewDevelopment()
			traceCfg := &jaegerConfig.Configuration{
				ServiceName: service,
				Sampler: &jaegerConfig.SamplerConfig{
					Type:  "const",
					Param: 100,
				},
				RPCMetrics: true,
			}

			tracer, _, err := traceCfg.NewTracer(
				jaegerConfig.Logger(jaegerZap.NewLogger(logger)),
			)
			if err != nil {
				logger.Fatal("failed to create tracer for service "+service, zap.Error(err))
			}

			fakeTracerStore.Store(service, tracer)

			return tracer, nil
		})

		return value.(opentracing.Tracer)
	}
)

func (w worker) simulateTraces() {
	tracer := opentracing.GlobalTracer()

	var i int
	for atomic.LoadUint32(w.running) == 1 {
		rootSpan := tracer.StartSpan("lets-go")
		ext.SpanKindRPCClient.Set(rootSpan)
		ext.PeerService.Set(rootSpan, "tracegen-service")
		ext.PeerHostIPv4.Set(rootSpan, fakeIP())
		if w.debug {
			ext.SamplingPriority.Set(rootSpan, 100)
		}

		if w.firehose {
			jaeger.EnableFirehose(rootSpan.(*jaeger.Span))
		}

		parentCtx := rootSpan.Context()
		if w.marshal {
			m := make(map[string]string)
			c := opentracing.TextMapCarrier(m)
			if err := tracer.Inject(rootSpan.Context(), opentracing.TextMap, c); err == nil {
				c := opentracing.TextMapCarrier(m)
				parentCtx, err = tracer.Extract(opentracing.TextMap, c)
				if err != nil {
					w.logger.Error("cannot extract from TextMap", zap.Error(err))
				}
			} else {
				w.logger.Error("cannot inject span", zap.Error(err))
			}
		}

		var (
			totalLatency   time.Duration
			chainedLatency time.Duration
			sleptLatency   time.Duration
			issuedAt       = time.Now()
		)
		for _, chainedService := range w.chainedServices {
			chainedLatency, sleptLatency = fakeTraces(parentCtx, strings.Split(chainedService, ","), w.serviceApis, chainedLatency)

			totalLatency += chainedLatency
			totalLatency += sleptLatency
		}

		rootSpan.FinishWithOptions(opentracing.FinishOptions{FinishTime: issuedAt.Add(totalLatency)})

		i++
		if w.traces != 0 {
			if i >= w.traces {
				break
			}
		}
	}
	w.logger.Info(fmt.Sprintf("Worker %d generated %d traces", w.id, i))
	w.wg.Done()
}

var (
	redisApis = []string{
		"Incr", "Set", "HSet", "TTL",
	}

	mysqlApis = []string{
		"Query", "Update", "Delete", "Update", "Replace",
	}
)

func fakeTraces(parentCtx opentracing.SpanContext, chainedServices []string, serviceApis []string, chainedLatency time.Duration) (totalLatency, sleptLatency time.Duration) {
	if chainedLatency > 0 {
		time.Sleep(chainedLatency)
	}

	tracer := fakeTracer(chainedServices[0])

	var (
		rootSpan     opentracing.Span
		rootIssuedAt = time.Now()
	)
	for i, chainedService := range chainedServices {
		fakeApi := serviceApis[rand.Intn(len(serviceApis))]
		fakeLatency := fakeSpanDuration()

		var isFakeDB bool
		switch {
		case strings.HasPrefix(chainedService, "redis-"):
			isFakeDB = true
			chainedService = "redis"
			fakeApi = redisApis[rand.Intn(len(redisApis))]

		case strings.HasPrefix(chainedService, "mysql-"):
			isFakeDB = true
			chainedService = "mysql"
			fakeApi = mysqlApis[rand.Intn(len(mysqlApis))]

		}

		var (
			childSpan     opentracing.Span
			childIssuedAt time.Time
			pausedLatency time.Duration
		)
		if !isFakeDB {
			if i > 0 {
				pausedLatency += fakePausedDuration(fakeLatency)

				childIssuedAt = time.Now()
				childSpan = tracer.StartSpan(fakeApi, opentracing.ChildOf(parentCtx))

				ext.SpanKindRPCClient.Set(childSpan)
			} else {
				childIssuedAt = time.Now()
				rootSpan = tracer.StartSpan(fakeApi, ext.RPCServerOption(parentCtx))

				childSpan = rootSpan
			}

			parentCtx = childSpan.Context()

			method2path := strings.SplitN(fakeApi, ":", 2)
			ext.HTTPMethod.Set(childSpan, method2path[0])
			ext.HTTPUrl.Set(childSpan, method2path[1])
		} else {
			if i > 0 {
				pausedLatency += fakePausedDuration(fakeLatency)
			}

			childIssuedAt = time.Now()
			childSpan = tracer.StartSpan(fakeApi, opentracing.ChildOf(parentCtx))

			ext.SpanKindRPCClient.Set(childSpan)
			ext.DBType.Set(childSpan, chainedService)
			ext.DBInstance.Set(childSpan, fake.IPv4())
			ext.DBStatement.Set(childSpan, fmt.Sprintf("%s * FROM `services` WHERE service=`%s`", strings.ToUpper(fakeApi), chainedService))
		}

		ext.PeerService.Set(childSpan, chainedService)
		ext.PeerHostIPv4.Set(childSpan, fakeIP())

		switch rn := rand.Intn(100); {
		case rn%11 == 0, rn%13 == 0, rn%17 == 0:
			if !isFakeDB {
				ext.HTTPStatusCode.Set(childSpan, uint16(500+rand.Intn(4)))
			}

			ext.LogError(childSpan, fmt.Errorf("invoke service %s with error", chainedService), log.String("trace-error", fake.WordsN(rn)))
		default:
			if !isFakeDB {
				ext.HTTPStatusCode.Set(childSpan, 200)
			}
		}

		if i > 0 {
			childSpan.FinishWithOptions(opentracing.FinishOptions{FinishTime: childIssuedAt.Add(fakeLatency)})
		}

		totalLatency += fakeLatency
		sleptLatency += pausedLatency
		if totalLatency < fakeLatency+pausedLatency {
			totalLatency = fakeLatency + pausedLatency
			sleptLatency += pausedLatency
		}
	}

	rootSpan.FinishWithOptions(opentracing.FinishOptions{FinishTime: rootIssuedAt.Add(totalLatency + sleptLatency)})

	return
}
