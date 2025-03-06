package krakend

import (
	"fmt"
	"net/http"

	botdetector "github.com/krakendio/krakend-botdetector/v2/gin"
	jose "github.com/krakendio/krakend-jose/v2"
	ginjose "github.com/krakendio/krakend-jose/v2/gin"
	lua "github.com/krakendio/krakend-lua/v2/router/gin"
	metrics "github.com/krakendio/krakend-metrics/v2/gin"
	opencensus "github.com/krakendio/krakend-opencensus/v2/router/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	router "github.com/luraproject/lura/v2/router/gin"

	"github.com/gin-gonic/gin"
)

// NewHandlerFactory returns a HandlerFactory with a rate-limit and a metrics collector middleware injected
func NewHandlerFactory(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory) router.HandlerFactory {
	handlerFactory := router.CustomErrorEndpointHandler(logger, TimeoutToHTTPError)
	handlerFactory = CustomTimeoutHandler(logger, handlerFactory)
	handlerFactory = lua.HandlerFactory(logger, handlerFactory)
	handlerFactory = ginjose.HandlerFactory(handlerFactory, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	handlerFactory = botdetector.New(handlerFactory, logger)

	return func(cfg *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		logger.Debug(fmt.Sprintf("[ENDPOINT: %s] Building the http handler", cfg.Endpoint))
		return handlerFactory(cfg, p)
	}
}

type handlerFactory struct{}

func (handlerFactory) NewHandlerFactory(l logging.Logger, m *metrics.Metrics, r jose.RejecterFactory) router.HandlerFactory {
	return NewHandlerFactory(l, m, r)
}

func CustomTimeoutHandler(l logging.Logger, next router.HandlerFactory) router.HandlerFactory {
	return func(remote *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		handlerFunc := next(remote, p)
		return func(c *gin.Context) {
			handlerFunc(c)

			if c.Writer.Status() == http.StatusRequestTimeout {
				message := `{"errorCode": "timeoutError", "errorMessage": "The request timed out"}`
				c.Writer.Header().Set("content-type", "application/json")
				_, err := c.Writer.WriteString(message)
				if err != nil {
					return
				}
			}
		}
	}
}

func TimeoutToHTTPError(err error) int {
	if err.Error() == "context deadline exceeded" {
		return http.StatusRequestTimeout
	}
	return http.StatusInternalServerError
}
