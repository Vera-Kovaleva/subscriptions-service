package log

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func ErrorAttr(err error) slog.Attr {
	return slog.String("error", err.Error())
}

func SetRequestID(ctx *gin.Context, id string) {
	ctx.Set("request_id", id)
}

func RequestID(ctx context.Context) slog.Attr {
	id := "unknown"

	if ginCtx, ok := ctx.(*gin.Context); ok {
		if reqID := ginCtx.GetString("request_id"); reqID != "" {
			id = reqID
		}
	}

	if val := ctx.Value("request_id"); val != nil {
		if reqID, ok := val.(string); ok && reqID != "" {
			id = reqID
		}
	}

	return slog.String("request_id", id)
}
