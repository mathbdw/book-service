package app

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/pressly/goose/v3"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/mathbdw/book/config"
	repo_kafka "github.com/mathbdw/book/internal/infrastructure/kafka"
	repo_observability "github.com/mathbdw/book/internal/infrastructure/observability"
	impmetric "github.com/mathbdw/book/internal/infrastructure/observability/opentelemetry/metrics"
	imptracer "github.com/mathbdw/book/internal/infrastructure/observability/opentelemetry/tracers"
	book_repo "github.com/mathbdw/book/internal/infrastructure/persistence/postgres"
	book_grpc_handler "github.com/mathbdw/book/internal/interfaces/controllers/grpc/v1/handlers"
	status_controller "github.com/mathbdw/book/internal/interfaces/controllers/status"
	book_bot_handler "github.com/mathbdw/book/internal/interfaces/controllers/telegram_bot/v1/handlers"
	"github.com/mathbdw/book/internal/interfaces/observability"
	book_usecase "github.com/mathbdw/book/internal/usecases/book"
	uc_services "github.com/mathbdw/book/internal/usecases/services"
	"github.com/mathbdw/book/pkg/gateway"
	"github.com/mathbdw/book/pkg/grpcserver"
	pkg_producer "github.com/mathbdw/book/pkg/kafka/producer"
	pkg_logger "github.com/mathbdw/book/pkg/logger/zerolog"
	pkg_metric "github.com/mathbdw/book/pkg/metric/opentelemetry"
	pkg_postgres "github.com/mathbdw/book/pkg/postgres"
	status_server "github.com/mathbdw/book/pkg/status"
	pkg_tbot "github.com/mathbdw/book/pkg/tbot"
	pkg_tracer "github.com/mathbdw/book/pkg/tracer/opentelemetry"
)

// initLogger - initializing logger
func initLogger(cfg *config.Config) observability.Logger {
	logger := pkg_logger.New(cfg)
	logger.Debug("app.initLogger: config running ", map[string]any{"config": cfg})

	return logger
}

// initPostgres - initializing postgres
func initPostgres(cfg *config.Config, logger observability.Logger) *pkg_postgres.Postgres {
	pg, err := pkg_postgres.New(
		logger,
		pkg_postgres.Dsn(cfg.Database),
		pkg_postgres.Driver(cfg.Database.Driver),
		pkg_postgres.MaxOpenConns(cfg.Database.MaxOpenConns),
		pkg_postgres.MaxIdleConns(cfg.Database.MaxIdleConns),
		pkg_postgres.ConnMaxIdleTime(cfg.Database.ConnMaxIdleTime),
		pkg_postgres.ConnMaxLifetime(cfg.Database.ConnMaxLifetime),
	)
	if err != nil {
		logger.Fatal("app.initPostgres: PG new", map[string]any{"error": err})
	}

	return pg
}

// applyMigration - apply migration
func applyMigration(cfg *config.Config, pg *pkg_postgres.Postgres, logger observability.Logger) {
	if err := goose.Up(pg.Sqlx.DB, cfg.Database.Migrations); err != nil {
		logger.Fatal("app.applyMigration: failed migration", map[string]any{"err": err})
	}
}

// initTracer - initializing tracer
func initTracer(ctx context.Context, cfg *config.Config, logger observability.Logger) *sdktrace.TracerProvider {
	tracer, err := pkg_tracer.New(
		ctx,
		cfg,
		pkg_tracer.WithAddress(cfg.Tracer.Host, cfg.Tracer.Port),
		pkg_tracer.WithCompressor("gzip"),
		pkg_tracer.WithTimeout(cfg.Tracer.Timeout),
		pkg_tracer.WithInsecure(cfg.Tracer.Insecure),
	)
	if err != nil {
		logger.Fatal("app.initTracer: tracer new", map[string]any{"err": err})

		return nil
	}

	tp, err := tracer.Start(ctx, cfg)
	if err != nil {
		logger.Fatal("app.initTracer: failed init", map[string]any{"err": err})
	}
	logger.Debug("app.initTracer: tracer started", nil)

	return tp
}

// initTracer - initializing metric
func initMetric(ctx context.Context, cfg *config.Config, logger observability.Logger) *sdkmetric.MeterProvider {
	metric, err := pkg_metric.New(
		ctx,
		cfg,
		pkg_metric.WithAddress(cfg.Metric.Host, cfg.Metric.Port),
		pkg_metric.WithCompressor("gzip"),
		pkg_metric.WithTimeout(cfg.Metric.Timeout),
		pkg_metric.WithInsecure(cfg.Metric.Insecure),
	)
	if err != nil {
		logger.Fatal("app.initMetric: metric new", map[string]any{"err": err})
	}

	mp, err := metric.Start(ctx, cfg)
	if err != nil {
		logger.Fatal("app.initMetric: failed init", map[string]any{"err": err})
	}
	logger.Debug("app.initMetric: metric started", nil)

	return mp
}

// initObservability - initializing observability
func initObservability(ctx context.Context, cfg *config.Config, tp *sdktrace.TracerProvider, mp *sdkmetric.MeterProvider, logger observability.Logger) *repo_observability.Observability {
	metricHandlers, err := impmetric.NewOpentelemetryHandlerMetrics(mp)
	if err != nil {
		logger.Error("app.initObservability: failed init metricHandlers", map[string]any{"err": err})
	}

	metricUsecases, err := impmetric.NewOpentelemetryUsecaseMetrics(mp)
	if err != nil {
		logger.Error("app.initObservability: failed init metricUsecases", map[string]any{"err": err})
	}

	metricRepositories, err := impmetric.NewOpentelemetryRepositoryMetrics(mp)
	if err != nil {
		logger.Error("app.initObservability: failed init metricRepositories", map[string]any{"err": err})
	}

	tracerHandlers, err := imptracer.NewOpentelemetryHandlerTracer(tp)
	if err != nil {
		logger.Error("app.initObservability: failed init tracerHandlers", map[string]any{"err": err})
	}

	tracerUsecases, err := imptracer.NewOpentelemetryUsecaseTracer(tp)
	if err != nil {
		logger.Error("app.initObservability: failed init tracerUsecases", map[string]any{"err": err})
	}

	tracerRepositories, err := imptracer.NewOpentelemetryRepositoryTracer(tp)
	if err != nil {
		logger.Error("app.initObservability: failed init tracerRepositories", map[string]any{"err": err})
	}

	observ, err := repo_observability.NewObservabilityBuilder().
		WithLogger(logger).
		WithHandlerMertic(metricHandlers).
		WithUsecaseMertic(metricUsecases).
		WithRepositoryMertic(metricRepositories).
		WithHandlerTracer(tracerHandlers).
		WithUsecaseTracer(tracerUsecases).
		WithRepositoryTracer(tracerRepositories).
		Build()

	if err != nil {
		logger.Error("app.initObservability: failed init observ", map[string]any{"err": err})
	}

	return observ
}

// RunPublisher - run publisher servic
func RunPublisher(cfg *config.Config) {
	var err error

	ctx := context.Background()

	logger := initLogger(cfg)
	pg := initPostgres(cfg, logger)
	defer pg.Sqlx.Close()

	tp := initTracer(ctx, cfg, logger)
	mp := initMetric(ctx, cfg, logger)
	observ := initObservability(ctx, cfg, tp, mp, logger)

	producerPkg := pkg_producer.New(
		pkg_producer.WithBrokers(cfg.Kafka.Brokers),
		pkg_producer.WithReturnSuccesses(cfg.Kafka.Producer.ReturnSuccesses),
		pkg_producer.WithRequiredAcks(cfg.Kafka.Producer.RequiredAcks),
		pkg_producer.WithCompression(cfg.Kafka.Producer.Compression),
		pkg_producer.WithPartitioner(cfg.Kafka.Producer.Partitioner),
	)
	producer, err := producerPkg.Start()
	if err != nil {
		logger.Fatal("app.RunPublisher: producer start", map[string]any{"error": err})
	}
	defer producer.Close()

	bookEventRepo := book_repo.NewBookEventRepository(pg.Sqlx, pg.Builder, observ.ForRepository())
	publisher, err := repo_kafka.New(cfg.Kafka.Topics.Publish, producer, logger)
	if err != nil {
		logger.Fatal("app.RunPublisher: init publisher", map[string]any{"err": err})
	}
	publisherUsecases := uc_services.New(
		bookEventRepo,
		publisher,
		logger,
		uc_services.WithBatchSize(cfg.Kafka.Publisher.BatchSize),
		uc_services.WithInterval(cfg.Kafka.Publisher.Interval),
		uc_services.WithCountWorkers(cfg.Kafka.Publisher.CountWorkers),
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- publisherUsecases.Start(ctx)
	}()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		logger.Error("app.RunPublisher: ", map[string]any{"signal": s.String()})
	case <-errCh:
		logger.Error("app.RunPublisher: publisher shutting down...", nil)
	}

}

// RunApp - run bot servic
func RunBot(cfg *config.Config) {

	ctx := context.Background()

	logger := initLogger(cfg)
	pg := initPostgres(cfg, logger)
	defer pg.Sqlx.Close()

	tp := initTracer(ctx, cfg, logger)
	mp := initMetric(ctx, cfg, logger)
	observ := initObservability(ctx, cfg, tp, mp, logger)

	bookRepo := book_repo.NewBookRepository(pg.Sqlx, pg.Builder, observ.ForRepository())
	uowRepo := book_repo.NewUnitOfWork(pg.Sqlx, pg.Builder, observ.ForRepository())
	addBookUC := book_usecase.NewAddBookUsecase(uowRepo, observ.ForUsecases())
	getBookUC := book_usecase.NewGetBookUsecase(bookRepo, observ.ForUsecases())
	listBookUC := book_usecase.NewListBookUsecase(bookRepo, observ.ForUsecases())
	removeBookUC := book_usecase.NewRemoveBookUsecase(uowRepo, observ.ForUsecases())

	uc := book_usecase.New(
		book_usecase.WithAddBookUsecase(addBookUC),
		book_usecase.WithGetBookUsecase(getBookUC),
		book_usecase.WithListBookUsecase(listBookUC),
		book_usecase.WithRemoveBookUsecase(removeBookUC),
	)

	bot, err := pkg_tbot.New(
		pkg_tbot.WithToken(cfg.Bot.Token),
		pkg_tbot.WithReadTimeout(cfg.Bot.ReadTimeout),
		pkg_tbot.WithDebug(cfg.Project.Debug),
	)
	if err != nil {
		logger.Fatal("app.RunBot: bot new", map[string]any{"error": err})
	}

	hb := book_bot_handler.New(bot.Api, uc, observ.ForHandler())
	for update := range bot.Start() {
		hb.Start(ctx, update)
	}

	logger.Info("app.RunBot: the bot service is ready to accept requests", nil)

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		logger.Error("app.RunBot", map[string]any{"signal": s.String()})
	}

	// Shutdown
	bot.Shutdown()
	logger.Error("app.RunBot: bot shutting down...", nil)
}

// RunApp - run app servic
func RunApp(cfg *config.Config) {
	var err error

	ctx := context.Background()

	logger := initLogger(cfg)
	pg := initPostgres(cfg, logger)
	defer pg.Sqlx.Close()

	applyMigration(cfg, pg, logger)
	tp := initTracer(ctx, cfg, logger)
	mp := initMetric(ctx, cfg, logger)
	observ := initObservability(ctx, cfg, tp, mp, logger)

	gatewayServer := gateway.New(
		gateway.Address(cfg.Rest.Host, cfg.Rest.Port),
		gateway.AddressGrpc(cfg.Grpc.Host, cfg.Grpc.Port),
	)
	grpcServer := grpcserver.New(
		grpcserver.Address(cfg.Grpc.Host, cfg.Grpc.Port),
		grpcserver.Mode(cfg.Project.Debug),
	)
	statusServer := status_server.New(
		status_server.Address(cfg.Status.Host, cfg.Status.Port),
	)

	isReady := &atomic.Value{}
	isReady.Store(false)
	status_controller.NewRouter(statusServer, cfg, isReady, logger)

	bookRepo := book_repo.NewBookRepository(pg.Sqlx, pg.Builder, observ.ForRepository())
	uowRepo := book_repo.NewUnitOfWork(pg.Sqlx, pg.Builder, observ.ForRepository())
	addBookUC := book_usecase.NewAddBookUsecase(uowRepo, observ.ForUsecases())
	getBookUC := book_usecase.NewGetBookUsecase(bookRepo, observ.ForUsecases())
	listBookUC := book_usecase.NewListBookUsecase(bookRepo, observ.ForUsecases())
	removeBookUC := book_usecase.NewRemoveBookUsecase(uowRepo, observ.ForUsecases())

	uc := book_usecase.New(
		book_usecase.WithAddBookUsecase(addBookUC),
		book_usecase.WithGetBookUsecase(getBookUC),
		book_usecase.WithListBookUsecase(listBookUC),
		book_usecase.WithRemoveBookUsecase(removeBookUC),
	)

	book_grpc_handler.NewBookHandler(
		grpcServer.App,
		uc,
		observ.ForHandler(),
	)

	// Start servers
	gatewayServer.Start(logger)
	grpcServer.Start()
	statusServer.Start()

	go func() {
		time.Sleep(2 * time.Second)
		isReady.Store(true)
		logger.Info("app.RunApp: the service is ready to accept requests", nil)
	}()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	isReady.Store(false)

	select {
	case s := <-interrupt:
		logger.Error("app.RunApp:", map[string]any{"signal": s.String()})
	case err = <-gatewayServer.Notify():
		logger.Error("app.RunApp: gatewayServer.Notify", map[string]any{"error": err.Error()})
	case err = <-grpcServer.Notify():
		logger.Error("app.RunApp: grpcServer.Notify", map[string]any{"error": err.Error()})
	case err = <-statusServer.Notify():
		logger.Error("app.RunApp: statusServer.Notify", map[string]any{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Shutdown
	tp.Shutdown(ctx)
	logger.Error("app.RunApp: traceProvider shutting down...", nil)
	mp.Shutdown(ctx)
	logger.Error("app.RunApp: metricProvider shutting down...", nil)

	err = gatewayServer.Shutdown(ctx)
	if err != nil {
		logger.Error("app.RunApp: gatewayServer.Shutdown", map[string]any{"error": err.Error()})
	} else {
		logger.Error("app.RunApp: gatewayServer shutting down...", nil)
	}

	grpcServer.Shutdown()
	logger.Error("app.RunApp: grpcServer shutting down...", nil)

	err = statusServer.Shutdown(ctx)
	if err != nil {
		logger.Error("app.RunApp: statusServer.Shutdown", map[string]any{"error": err.Error()})
	} else {
		logger.Error("app.RunApp: statusServer shutting down...", nil)
	}
}
