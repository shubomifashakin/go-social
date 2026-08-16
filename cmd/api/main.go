package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/db"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"github.com/shubomifashakin/go-social/internal/routers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	serviceName := "go-social"

    environment := os.Getenv("APP_ENV")
    if strings.TrimSpace(environment) == "" {
        godotenv.Load() 
        environment = os.Getenv("APP_ENV") 
    }

	// configuring zap logger
	config:= zap.NewProductionConfig()
	config.Development= false
	config.Level= zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.Encoding="json"
	config.Sampling= &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
	logger,err:=config.Build(
		zap.AddStacktrace(zap.ErrorLevel),
		zap.Fields(zap.String("service", serviceName), 
		zap.String("env", environment)),
	)
   
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

	// get the port from the env
	port:= os.Getenv("PORT")
	if _,err:=strconv.Atoi(port); err != nil {
		logger.Fatal("Invalid port specified",zap.String("port",port))	
	}

	resendApiKey:= os.Getenv("RESEND_API_KEY")
	if resendApiKey == "" {
		logger.Fatal("Resend is not configured")
	}

	// setup database
	databaseUrl:= os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		logger.Fatal("DATABASE is not configured")
	}

	redisUrl:= os.Getenv("REDIS_URL")
	if redisUrl == "" {
		logger.Fatal("REDIS is not configured")
	}

	fromMail:= os.Getenv("MAILER_FROM")
	if fromMail == ""{
		logger.Fatal("MAILER_FROM not set")
	}

	ctx,cancel:= context.WithTimeout(context.Background(),time.Second*30)
	defer cancel()

	// create the instance of the db
	db,err:= db.ConnectDb(ctx,databaseUrl)
	if err != nil {
		logger.Fatal("Failed to connect to db",zap.Error(err))
	}
	defer db.Close()
	
	// create instance of cache
	redisInstance,err:= cache.NewRedisClient(ctx,redisUrl,serviceName)
	if err != nil {
		logger.Fatal("Failed to connect to redis",zap.Error(err))
	}
	defer redisInstance.Close()

	// create the instance of the mailer
	resendInstance:= mailer.NewMailer(resendApiKey)

	mux := routers.RegisterRouter(db, redisInstance, resendInstance, logger, fromMail)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3001"},
		AllowCredentials: true,
		AllowedMethods: []string{"GET","POST","PUT","PATCH","DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization","Content-Type"},
		Debug: environment == "",
	})
	
	server := &http.Server{
		Handler: c.Handler(mux),
		Addr: ":"+port,
		WriteTimeout: 15*time.Second,
		ReadTimeout: 15*time.Second,
		IdleTimeout: 30*time.Second,
	}

	// listen to signals for the process
	sigChan:= make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// graceful shutdown
	go func(){
		<- sigChan
		logger.Info("Shutting down server")

		// server should shutdown within 15 seconds
		ctx,cancel:=context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err:=server.Shutdown(ctx)
		if err !=nil {
			logger.Error("Forced shutdown:", zap.Error(err))
			return
		}

		logger.Info("Server shutdown gracefully")
	}()

	logger.Info("Starting server",zap.String("port", port))

	err=server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}