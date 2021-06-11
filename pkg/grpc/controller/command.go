package controller

import (
	"github.com/abergasov/go_telegram_bots/pkg/utils"
	"log"

	"go.uber.org/zap"

	"github.com/abergasov/go_telegram_bots/pkg/logger"

	"github.com/abergasov/go_telegram_bots/pkg/config"
	pb "github.com/abergasov/go_telegram_bots/pkg/grpc"
)

type ServerController struct {
	bot CommandBot
}

func InitServerController(cnf *config.AppConfig, bot CommandBot) *ServerController {
	return &ServerController{
		bot: bot,
	}
}

func (s ServerController) ListenCommands(req *pb.Request, srv pb.CommandStream_ListenCommandsServer) error {
	streamID := utils.RandString(10)
	log.Println("start new server: " + streamID)
	var resp pb.Response
	for cmd := range s.bot.GetControlChan(req.TargetChat) {
		resp = pb.Response{Cmd: cmd.Cmd, ActionID: cmd.ActionID}
		if err := srv.Send(&resp); err != nil {
			logger.Error("error send command: "+streamID, err)
			return err
		}
		logger.Info("send command: "+streamID, zap.String("cmd", cmd.Cmd), zap.String("action", cmd.ActionID), zap.Int64("chat", req.TargetChat))
	}
	return nil
}
