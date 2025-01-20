package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/timmbarton/errors"
	"github.com/timmbarton/utils/tracing"
	"go.opentelemetry.io/otel/attribute"
)

const TraceIdHeader = "X-Trace-Id"

type response struct {
	TraceID string    `json:"trace_id"`
	Error   *errs.Err `json:"error"`
}

func setErrCode(err *errs.Err, httpStatusCode int) {
	switch httpStatusCode {
	case http.StatusBadRequest:
		err.Code = errs.ErrCodeBadRequest
	case http.StatusUnauthorized:
		err.Code = errs.ErrCodeUnauthorized
	case http.StatusForbidden:
		err.Code = errs.ErrCodeForbidden
	case http.StatusNotFound:
		err.Code = errs.ErrCodeNotFound
	case http.StatusMethodNotAllowed:
		err.Code = errs.ErrCodeNotAllowed
	case http.StatusRequestTimeout:
		err.Code = errs.ErrCodeRequestTimeout
	case http.StatusInternalServerError:
		err.Code = errs.ErrCodeInternal
	case http.StatusNotImplemented:
		err.Code = errs.ErrCodeNotImplemented
	case http.StatusBadGateway:
		err.Code = errs.ErrCodeBadGateway
	default:
		err.Code = errs.ErrCodeUnknown
	}
}

func GetErrsMiddleware(
	serviceId int,
	showUnknownErrorsInResponse bool,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, span := tracing.NewSpan(c.UserContext())
		defer span.End()

		c.SetUserContext(ctx)
		span.SetName("errs.GetErrsMiddleware")
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
				err = &errs.Err{
					Code:    0,
					Index:   0,
					Message: fiberErr.Message,
				}
				setErrCode(err, fiberErr.Code)
			} else {
				customErr, ok := errs.Parse(errInterface)
				if ok {
					err = customErr
				} else if showUnknownErrorsInResponse {
					err = &errs.Err{
						Code:    errs.ErrCodeInternal,
						Index:   serviceId * 1_0000,
						Message: fmt.Sprintf("на проде этого сообщения не будет: %s", errInterface.Error()),
					}
				} else {
					err = errs.ErrUnknown
				}
			}
		} else if err.Code >= errs.ErrCodeInternal && !showUnknownErrorsInResponse {
			err.Index = serviceId * 1_0000
			err.Message = "Внутренняя ошибка сервера"
			err.Params = nil
		}

		resp.Error = err
		if resp.Error.Code == 0 {
			resp.Error = errs.ErrUnknown
		}

		// logging

		respJSON := ""

		respJSONBytes, marshallingErr := json.Marshal(resp)
		if marshallingErr != nil {
			respJSON = fmt.Sprintf("%#+v (marshalling err: %s)", resp, marshallingErr.Error())
		} else {
			respJSON = string(respJSONBytes)
		}

		// log error
		log.Printf(
			"ip: %s | path: %s | response: %s\n",
			c.IP(),
			c.Path(),
			respJSON,
		)

		return c.Status(int(resp.Error.Code)).JSON(resp)
	}

}
