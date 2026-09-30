package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/omkod2025/boiler-plate-go/configs"
	amqpdelivery "github.com/omkod2025/boiler-plate-go/delivery/amqp"
	"github.com/omkod2025/boiler-plate-go/delivery/http/middleware"
	"github.com/omkod2025/boiler-plate-go/delivery/receiver"
	"github.com/omkod2025/boiler-plate-go/delivery/sandbox"
	"github.com/omkod2025/boiler-plate-go/delivery/websocket"
	"github.com/omkod2025/boiler-plate-go/migrations"
	"github.com/omkod2025/boiler-plate-go/pkg/amqp"
	"github.com/omkod2025/boiler-plate-go/pkg/health"
	"github.com/omkod2025/boiler-plate-go/pkg/logger"
	"github.com/omkod2025/boiler-plate-go/pkg/migrate"
)

// Role ของ process: image เดียว (Dockerfile เดียว) รันได้หลายหน้าที่ เลือกด้วย flag -role หรือ APP_ROLE
//
//	api       HTTP API หลัก (DB + JWT)
//	stream    HTTP สำหรับ connection ยาว เช่น SSE/WebSocket (ไม่มี write timeout)
//	worker    consume RabbitMQ (AMQP_URL + AMQP_QUEUE) และงานเบื้องหลัง; มี health บน APP_PORT
//	receiver  รับ webhook: ไม่มี DB ตรวจแล้ว publish เข้า RabbitMQ ก่อนตอบ 2xx
//	sandbox   ประมวลผลงานเสี่ยง (เช่น parse ไฟล์) ไม่มี DB และไม่ควรมี network: อ่าน stdin เขียน stdout
//	migrate   รัน goose migration แล้วจบ (รันก่อน deploy role อื่น)
const (
	RoleAPI      = "api"
	RoleStream   = "stream"
	RoleWorker   = "worker"
	RoleReceiver = "receiver"
	RoleSandbox  = "sandbox"
	RoleMigrate  = "migrate"
)

// Roles ทั้งหมดที่รองรับ
var Roles = []string{RoleAPI, RoleStream, RoleWorker, RoleReceiver, RoleSandbox, RoleMigrate}

// Run เริ่ม role ตาม cfg.Env.APP_ROLE และบล็อกจน ctx ถูก cancel (SIGINT/SIGTERM) หรือเกิด error
func Run(ctx context.Context, cfg *configs.Config) error {
	switch cfg.Env.APP_ROLE {
	case RoleAPI:
		return runAPI(ctx, cfg)
	case RoleStream:
		return runStream(ctx, cfg)
	case RoleWorker:
		return runWorker(ctx, cfg)
	case RoleReceiver:
		return runReceiver(ctx, cfg)
	case RoleSandbox:
		return runSandbox(ctx, os.Stdin, os.Stdout)
	case RoleMigrate:
		return runMigrate(ctx, cfg)
	}
	return fmt.Errorf("unknown role %q (want one of %v)", cfg.Env.APP_ROLE, Roles)
}

func runAPI(ctx context.Context, cfg *configs.Config) error {
	if err := cfg.RequireJWT(); err != nil {
		return err
	}
	if err := cfg.ConnectDB(ctx); err != nil {
		return err
	}
	defer cfg.DB.Close()
	r := InitApp(ctx, cfg)
	if r == nil {
		return errors.New("failed to initialize application")
	}
	health.Register(r, dbCheck(cfg))
	return serveHTTP(ctx, cfg, &http.Server{Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second})
}

func runStream(ctx context.Context, cfg *configs.Config) error {
	if err := cfg.RequireJWT(); err != nil {
		return err
	}
	if err := cfg.ConnectDB(ctx); err != nil {
		return err
	}
	defer cfg.DB.Close()
	jwtConfig, err := loadJWTConfig(cfg.Env)
	if err != nil {
		return err
	}
	r := baseRouter(cfg)
	health.Register(r, dbCheck(cfg))
	g := r.Group(cfg.Env.APP_PREFIX)
	g.Use(middleware.JWTOptionalWithIgnoreRules(jwtConfig, nil))
	websocket.RegisterRoutes(g)
	// ไม่มี WriteTimeout: connection ของ SSE/WebSocket อยู่ได้นาน — handler ต้องดู ctx เองเพื่อปิดตอน shutdown
	return serveHTTP(ctx, cfg, &http.Server{Handler: r, ReadHeaderTimeout: 5 * time.Second})
}

func runWorker(ctx context.Context, cfg *configs.Config) error {
	if err := cfg.ConnectDB(ctx); err != nil {
		return err
	}
	defer cfg.DB.Close()
	checks := []health.Check{dbCheck(cfg)}
	errs := make(chan error, 2)
	if cfg.Env.AMQP_URL != "" {
		if cfg.Env.AMQP_QUEUE == "" {
			return errors.New("worker: AMQP_QUEUE is required when AMQP_URL is set")
		}
		conn, err := amqp.Dial(cfg.Env.AMQP_URL, cfg.Env.APP_NAME+"-worker")
		if err != nil {
			return err
		}
		defer func() { _ = conn.Close() }()
		checks = append(checks, health.Check{Name: "amqp", Essential: true, Probe: conn.Ping})
		d := dependencies{cfg: cfg, db: cfg.DB}
		consumer := &amqp.Consumer{Conn: conn, Queue: cfg.Env.AMQP_QUEUE, Prefetch: cfg.Env.AMQP_PREFETCH, Handler: newWorkerHandler(d)}
		go func() { errs <- consumer.Run(ctx) }()
		logger.Info("worker: consuming " + cfg.Env.AMQP_QUEUE)
	} else {
		logger.Warn("worker: AMQP_URL not set; only background jobs and health run")
	}
	r := baseRouter(cfg)
	health.Register(r, checks...)
	go func() { errs <- serveHTTP(ctx, cfg, &http.Server{Handler: r, ReadHeaderTimeout: 5 * time.Second}) }()
	return <-errs
}

func runReceiver(ctx context.Context, cfg *configs.Config) error {
	var pub receiver.Publisher
	var checks []health.Check
	if cfg.Env.AMQP_URL != "" {
		conn, err := amqp.Dial(cfg.Env.AMQP_URL, cfg.Env.APP_NAME+"-receiver")
		if err != nil {
			return err
		}
		defer func() { _ = conn.Close() }()
		pub = &amqp.Publisher{Conn: conn}
		checks = append(checks, health.Check{Name: "amqp", Essential: true, Probe: conn.Ping})
	} else {
		logger.Warn("receiver: AMQP_URL not set; webhooks are accepted and logged only")
	}
	r := baseRouter(cfg)
	health.Register(r, checks...)
	receiver.RegisterRoutes(r.Group(cfg.Env.APP_PREFIX), pub, cfg.Env.AMQP_EXCHANGE)
	return serveHTTP(ctx, cfg, &http.Server{Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second})
}

// runSandbox อ่านงานทีละบรรทัด (JSON) จาก in แล้วเขียนผลทีละบรรทัดลง out จนหมด input หรือ ctx ถูก cancel
func runSandbox(ctx context.Context, in io.Reader, out io.Writer) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, 64<<10), 16<<20)
	w := bufio.NewWriter(out)
	defer func() { _ = w.Flush() }()
	for sc.Scan() {
		if ctx.Err() != nil {
			return nil
		}
		result := sandbox.Process(ctx, sc.Bytes())
		if _, err := w.Write(append(result, '\n')); err != nil {
			return err
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return sc.Err()
}

func runMigrate(ctx context.Context, cfg *configs.Config) error {
	if err := cfg.ConnectDB(ctx); err != nil {
		return err
	}
	defer cfg.DB.Close()
	v, err := migrate.Up(ctx, cfg.DB.GetDB(), migrations.FS)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("migrate: database at version %d", v))
	return nil
}

func baseRouter(cfg *configs.Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Tracing(cfg.Env.APP_NAME), middleware.Recovery())
	return r
}

func dbCheck(cfg *configs.Config) health.Check {
	return health.Check{Name: "db", Essential: true, Probe: func(ctx context.Context) error { return cfg.DB.GetDB().Ping(ctx) }}
}

// serveHTTP รัน server บน APP_PORT จน ctx ถูก cancel แล้วปิดแบบรอ request ที่ค้างอยู่ (SHUTDOWN_TIMEOUT)
func serveHTTP(ctx context.Context, cfg *configs.Config, srv *http.Server) error {
	srv.Addr = ":" + cfg.Env.APP_PORT
	errs := make(chan error, 1)
	go func() {
		logger.Info(cfg.Env.APP_ROLE + " listening on :" + cfg.Env.APP_PORT)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
		close(errs)
	}()
	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Env.SHUTDOWN_TIMEOUT)
	defer cancel()
	logger.Info(cfg.Env.APP_ROLE + ": shutting down")
	return srv.Shutdown(shutdownCtx)
}

// newWorkerHandler ประกอบ handler ของ worker — เพิ่ม dependency ของ use case ที่ worker ใช้ที่นี่
func newWorkerHandler(d dependencies) amqp.Handler {
	return amqpdelivery.NewHandler()
}
