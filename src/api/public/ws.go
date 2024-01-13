package publicapi

import (
	"errors"

	"github.com/CE-Thesis-2023/backend/src/biz/handlers"
	"github.com/CE-Thesis-2023/backend/src/biz/service"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func WsAuthorizeLtd() handlers.AuthorizerHandler {
	return func(ctx *fiber.Ctx) error {
		deviceId := ctx.Params("id")
		if len(deviceId) == 0 {
			return custerror.FormatPermissionDenied("missing deviceId as parameter")
		}
		devices, err := service.GetWebService().GetDevices(ctx.Context(), &web.GetTranscodersRequest{
			Ids: []string{deviceId},
		})
		if err != nil {
			if errors.Is(err, custerror.ErrorNotFound) {
				logger.SDebug("WsAuthorizeLtd: deviceId not found")
				return custerror.ErrorPermissionDenied
			}
			logger.SError("WsAuthorizeLtd: error", zap.Error(err))
			return err
		}

		if len(devices.Transcoders) == 0 {
			logger.SDebug("WsAuthorizeLtd: deviceId not found")
			return custerror.ErrorPermissionDenied
		}

		ltd := devices.Transcoders[0]
		logger.SDebug("WsAuthorizeLtd: authorizing LTD",
			zap.String("deviceId", ltd.DeviceId),
			zap.Any("device", ltd))
		return nil
	}
}

func WsListenToMessages() handlers.ConnectionHandler {
	return func(i *handlers.ConnectionInformation) error {
		logger.SInfo("WsListenToMessages: started",
			zap.String("deviceId", i.Id),
			zap.String("version", i.Version))
		conn := i.Connection
		var msg handlers.RequestReplyResponse
		for {
			if err := conn.ReadJSON(&msg); err != nil {
				logger.SError("WsListenToMessages: ReadJSON error", zap.Error(err))
				continue
			}
			logger.SDebug("WsListenToMessages: message received",
				zap.String("deviceId", i.Id))

			if msg.MessageId != 0 {
				logger.SDebug("WsListenToMessages: message is request reply")
				if err := i.RequestReply.OnResponse(&msg); err != nil {
					logger.SError("WsListenToMessages: RRCommunicator.OnResponse error",
						zap.Error(err))
					continue
				}
				logger.SDebug("WsListenToMessages: RRCommunicator.OnResponse completed")
			}

			logger.SDebug("WsListenToMessages: RRCommunicator message is not request reply")
		}
	}
}
