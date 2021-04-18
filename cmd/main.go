package main

import (
	"log"
	"net"
	"net/http"

	"github.com/abergasov/go_telegram_bots/pkg/config"
	"github.com/abergasov/go_telegram_bots/pkg/logger"
	"github.com/abergasov/go_telegram_bots/pkg/middleware"
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

type server struct {
}

func (s server) ListenCommands(req *pb.Request, srv pb.CommandStream_ListenCommandsServer) error {
	log.Println("start new server")
	var max int32
	ctx := srv.Context()
	max = 999
	for {

		// exit if context is done
		// or continue
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive data from stream
		if req.TargetChat <= max {
			continue
		}

		resp := pb.Response{Cmd: "12312", ActionID: "adwdawd"}
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
		log.Printf("send new max=%d", max)
	}
}

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

	authorized := router.GinEngine.Group("/bot/pi/client/")
	authorized.Use(middleware.AuthControllerMiddleware(appConf.KeyToken))
	authorized.GET("ws", func(c *gin.Context) {
		machineIP := c.GetHeader("X-Real-Ip")
		router.HandleClient(c.Writer, c.Request, machineIP)
	})
	authorized.POST("re", router.HandleClientRest)

	logger.Info("Starting server at port " + appConf.AppPort)

	errR := router.GinEngine.Run(":" + appConf.AppPort)
	if errR != nil {
		logger.Fatal("Common server error", errR)
	}

	// create listiner
	lis, err := net.Listen("tcp", ":50105")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCommandStreamServer(s, server{})
	//s.RegisterService(s, server{})
	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
