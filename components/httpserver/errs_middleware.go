package httpserver

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/timmbarton/errors"
	"github.com/timmbarton/utils/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/timmbarton/layout/log"
)

const TraceIdHeader = "X-Trace-Id"

type response struct {
	TraceID string    `json:"trace_id"`
	Error   *errs.Err `json:"error"`
}

func getErrByHTTPCode(httpStatusCode int, index int, message string) *errs.Err {
	errCode := errs.ErrCodeUnknown
	switch httpStatusCode {
	case http.StatusBadRequest:
		errCode = errs.ErrCodeBadRequest
	case http.StatusUnauthorized:
		errCode = errs.ErrCodeUnauthorized
	case http.StatusForbidden:
		errCode = errs.ErrCodeForbidden
	case http.StatusNotFound:
		errCode = errs.ErrCodeNotFound
	case http.StatusMethodNotAllowed:
		errCode = errs.ErrCodeNotAllowed
	case http.StatusRequestTimeout:
		errCode = errs.ErrCodeRequestTimeout
	case http.StatusInternalServerError:
		errCode = errs.ErrCodeInternal
	case http.StatusNotImplemented:
		errCode = errs.ErrCodeNotImplemented
	case http.StatusBadGateway:
		errCode = errs.ErrCodeBadGateway
	default:
		errCode = errs.ErrCodeUnknown
	}

	return errs.New(errCode, index, message)
}

func GetErrsMiddleware(
	serviceId int,
	showUnknownErrorsInResponse bool,
	logger *zap.Logger,
) fiber.Handler {
	l := log.NewWrappedLogger(logger)

	return func(c *fiber.Ctx) error {
		ctx, span := tracing.NewSpan(c.UserContext())
		defer span.End()

		c.SetUserContext(ctx)
		span.SetName("ErrsMiddleware")
		c.Set(TraceIdHeader, tracing.GetTraceID(span))

		errInterface := c.Next()
		if errInterface != nil {
			span.SetAttributes(attribute.String("error", errInterface.Error()))
		}

		if errInterface == nil {
			return nil
		}

		resp := response{
			TraceID: tracing.GetTraceID(span),
		}

		// parse error

		err := (*errs.Err)(nil)
		isCustomErr := errors.As(errInterface, &err)
		if !isCustomErr {
			fiberErr := (*fiber.Error)(nil)
			isFiberErr := errors.As(errInterface, &fiberErr)
			if isFiberErr {
				err = getErrByHTTPCode(fiberErr.Code, serviceId*1_0000, fiberErr.Message)
			} else {
				customErr, ok := errs.Parse(errInterface)
				if ok {
					err = customErr
				} else if showUnknownErrorsInResponse {
					err = errs.New(
						errs.ErrCodeInternal,
						serviceId*1_0000,
						fmt.Sprintf("на проде этого сообщения не будет: %s", errInterface.Error()),
					)
				} else {
					err = errs.ErrUnknown
				}
			}
		} else if err.GetCode() >= int(errs.ErrCodeInternal) && !showUnknownErrorsInResponse {
			err = errs.New(errs.ErrCodeInternal, serviceId*1_0000, "Внутренняя ошибка сервера")
		}

		resp.Error = err
		if resp.Error.GetCode() == 0 {
			resp.Error = errs.ErrUnknown
		}

		// logging

		l.Error(
			ctx,
			fmt.Sprintf("%s | %d | %v", c.Path(), resp.Error.GetCode(), err),
			zap.String("ip", c.IP()),
			zap.String("path", c.Path()),
			zap.Error(err),
			log.Json("response", resp),
		)

		return c.Status(resp.Error.GetCode()).JSON(resp)
	}

}
