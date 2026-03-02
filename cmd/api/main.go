package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	grpcUserServer "backend-challenge/internal/adapters/grpc"
	pb "backend-challenge/internal/adapters/grpc/proto"
	httpAdapter "backend-challenge/internal/adapters/http"
	mongoAdapter "backend-challenge/internal/adapters/repository/mongo"
	tokenAdapter "backend-challenge/internal/adapters/token"
	"backend-challenge/internal/adapters/worker"
	"backend-challenge/internal/config"
	"backend-challenge/internal/core/services"

	"google.golang.org/grpc"
)

//	@title			Backend Challenge API
//	@version		1.0
//	@description	This is a backend challenge API.
//	@contact.name	iBallbar
//	@contact.url	https://github.com/iballbar
//	@contact.email	bestballs@gmail.com

//	@host		localhost:8080
//	@BasePath	/

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mongoClient, err := mongoAdapter.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(ctx)

	userRepo := mongoAdapter.NewUserRepository(mongoClient, cfg.MongoDB)
	userService := services.NewService(userRepo)
	tokenService := tokenAdapter.NewTokenService(cfg.JWTSecret)

	lotteryRepo := mongoAdapter.NewLotteryRepository(mongoClient, cfg.MongoDB, cfg.LotteryReservationTTL)
	lotteryService := services.NewLotteryService(lotteryRepo)

	userHandler := httpAdapter.NewUserHandler(userService, tokenService)
	lotteryHandler := httpAdapter.NewLotteryHandler(lotteryService)
	router := httpAdapter.NewRouter(userHandler, lotteryHandler, tokenService)
	server := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: router,
	}

	// user count logger worker
	wg.Go(func() {
		worker.StartUserCountLogger(ctx, userRepo, cfg.UserCountInterval)
	})

	// gRPC Server
	fmt.Printf("gRPC Port: %s\n", cfg.GRPCPort)
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	userServer := grpcUserServer.NewServer(userService)
	pb.RegisterUserServiceServer(grpcServer, userServer)
	go func() {
		startMsg := fmt.Sprintf("Starting gRPC server on %s", cfg.GRPCPort)
		slog.Info(startMsg)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		slog.Info("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
		grpcServer.GracefulStop()
		cancel()
		wg.Wait()
		slog.Info("Shutdown complete.")
	}()

	startMsg := fmt.Sprintf("Starting HTTP server on %s", cfg.HTTPPort)
	slog.Info(startMsg)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
