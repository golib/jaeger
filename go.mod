module github.com/jaegertracing/jaeger

go 1.16

replace (
	github.com/apache/thrift => github.com/jaegertracing/thrift v1.13.0-patch1
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
)

require (
	github.com/ClickHouse/clickhouse-go v1.4.3
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/Shopify/sarama v1.28.0
	github.com/apache/thrift v0.13.0
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/corpix/uarand v0.1.1 // indirect
	github.com/crossdock/crossdock-go v0.0.0-20160816171116-049aabb0122b
	github.com/dgraph-io/badger v1.6.2
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-openapi/errors v0.20.0
	github.com/go-openapi/loads v0.20.2
	github.com/go-openapi/runtime v0.19.26
	github.com/go-openapi/spec v0.20.3
	github.com/go-openapi/strfmt v0.20.0
	github.com/go-openapi/swag v0.19.14
	github.com/go-openapi/validate v0.20.2
	github.com/gocql/gocql v0.0.0-20210129204804-4364a4b9cfdd
	github.com/gogo/googleapis v1.4.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-hclog v0.15.0
	github.com/hashicorp/go-plugin v1.4.0
	github.com/icrowley/fake v0.0.0-20180203215853-4178557ae428
	github.com/kr/pretty v0.2.1
	github.com/mjibson/esc v0.2.0
	github.com/olivere/elastic v6.2.35+incompatible
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.9.0
	github.com/rs/cors v1.7.0
	github.com/securego/gosec/v2 v2.6.1
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	github.com/vektra/mockery v1.1.2
	github.com/wadey/gocovmerge v0.0.0-20160331181800-b5bfa59ec0ad
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	go.uber.org/atomic v1.7.0
	go.uber.org/automaxprocs v1.4.0
	go.uber.org/zap v1.16.0
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210225134936-a50acf3fe073
	google.golang.org/grpc v1.36.0
	google.golang.org/grpc/examples v0.0.0-20210225234839-9dfe677337d5 // indirect
	gopkg.in/yaml.v2 v2.4.0
	honnef.co/go/tools v0.1.2
)
