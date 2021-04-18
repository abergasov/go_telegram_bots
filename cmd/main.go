package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc/keepalive"

	"go.uber.org/zap"

	"github.com/abergasov/go_telegram_bots/pkg/grpc/controller"

	"github.com/abergasov/go_telegram_bots/pkg/config"
	"github.com/abergasov/go_telegram_bots/pkg/logger"
	"github.com/abergasov/go_telegram_bots/pkg/routes"

	pb "github.com/abergasov/go_telegram_bots/pkg/grpc"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

var (
	hookPostfix = "/bot/"
	buildTime   = "_dev"
	buildHash   = "_dev"
)

func main() {
	logger.NewLogger()
	appConf := config.InitConf("/configs/conf.yaml")
	router := routes.InitRouter(appConf, hookPostfix, buildTime, buildHash)
	for i := range router.BotManager.HooksURLList {
		// для каждого бота регаем свой обработчик хуков, исходя из того постфикса, что определен у бота
		router.GinEngine.POST(i, router.HandleHook)
	}
	router.GinEngine.GET("/bot/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ok":         true,
			"build_time": buildTime,
			"build_hash": buildHash,
		})
	})

	controllerBot := router.BotManager.GetBot("@rmcpi_bot")
	if controllerBot != nil {
		server := controller.InitServerController(appConf, controllerBot.(controller.CommandBot))
		go startGRPC(appConf.GRPCPort, server)
	}

	logger.Info("Starting server at port " + appConf.AppPort)
	if errR := router.GinEngine.Run(":" + appConf.AppPort); errR != nil {
		logger.Fatal("Common server error", errR)
	}

}

func startGRPC(port string, server pb.CommandStreamServer) {
	logger.Info("start listen grpc on port", zap.String("port", port))
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		}),
	)
	pb.RegisterCommandStreamServer(s, server)
	if err = s.Serve(lis); err != nil {
		logger.Fatal("failed serve grpc", err)
	}
}
