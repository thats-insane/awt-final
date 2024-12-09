package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/thats-insane/awt-final/internal/data"
	"github.com/thats-insane/awt-final/internal/mailer"
)

const appVersion = "1.0.0"

type serverConfig struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type appDependencies struct {
	config      serverConfig
	logger      *slog.Logger
	userModel   data.UserModel
	bookModel   data.BookModel
	reviewModel data.ReviewModel
	listModel   data.ListModel
	tokenModel  data.TokenModel
	mailer      mailer.Mailer
	wg          sync.WaitGroup
}

func openDB(settings serverConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func main() {
	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "Server Port")
	flag.StringVar(&settings.env, "env", "development", "Environment(Development|Staging|Production)")
	flag.StringVar(&settings.db.dsn, "db-dsn", "postgres://bookclub:password@localhost/bookclub?sslmode=disable", "PostgreSQL DSN")
	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")
	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")
	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.StringVar(&settings.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&settings.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&settings.smtp.username, "smtp-username", "6b4f4a0f1e81c5", "SMTP username")
	flag.StringVar(&settings.smtp.password, "smtp-password", "7a8cb475eeb545", "SMTP password")
	flag.StringVar(&settings.smtp.sender, "smtp-sender", "Book Club Community <no-reply@commentscommunity.2021154337.net>", "SMTP sender")
	flag.Func("cors-trusted-origins", "Trusted CORS origins", func(s string) error {
		settings.cors.trustedOrigins = strings.Fields(s)
		return nil
	})
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	appInstance := &appDependencies{
		config:      settings,
		logger:      logger,
		userModel:   data.UserModel{DB: db},
		bookModel:   data.BookModel{DB: db},
		reviewModel: data.ReviewModel{DB: db},
		listModel:   data.ListModel{DB: db},
		tokenModel:  data.TokenModel{DB: db},
		mailer:      mailer.New(settings.smtp.host, settings.smtp.port, settings.smtp.username, settings.smtp.password, settings.smtp.sender),
	}

	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
