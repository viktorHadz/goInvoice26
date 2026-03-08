package logging

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/traceid"
	"github.com/golang-cz/devslog"
	"github.com/viktorHadz/goInvoice26/internal/config"
)

// InitLogger builds the application logger and the matching httplog options
// Dev/localhost => colored console logs + DEBUG
// Prod          => JSON logs + INFO
func InitLogger(cfg config.Config) (*slog.Logger, *httplog.Options) {
	isDev := cfg.Env == "localhost" || cfg.Env == "dev"

	level := slog.LevelInfo
	if isDev {
		level = slog.LevelDebug
	}

	schema := httplog.SchemaECS.Concise(isDev)

	handlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return schema.ReplaceAttr(groups, a)
		},
	}

	logger := slog.New(newHandler(isDev, handlerOpts))
	// Will attach if debuging env
	// .With(
	// slog.String("env", cfg.Env),
	// )

	if !isDev {
		logger = logger.With(
			slog.String("app", "invoicer"),
			slog.String("version", "v1.4.0"),
		)
	}

	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(level)

	httpLogOpts := &httplog.Options{
		Level:         level,
		Schema:        schema,
		RecoverPanics: true,

		Skip: func(req *http.Request, status int) bool {
			return status == http.StatusNotFound || status == http.StatusMethodNotAllowed
		},

		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: nil,

		LogRequestBody:  isDebugHeaderSet,
		LogResponseBody: isDebugHeaderSet,

		LogExtraAttrs: func(req *http.Request, reqBody string, status int) []slog.Attr {
			if status == http.StatusBadRequest || status == http.StatusUnprocessableEntity {
				req.Header.Del("Authorization")
				return []slog.Attr{
					slog.String("curl", httplog.CURL(req, reqBody)),
				}
			}
			return nil
		},
	}

	return logger, httpLogOpts
}

func newHandler(isDev bool, opts *slog.HandlerOptions) slog.Handler {
	if isDev {
		base := devslog.NewHandler(os.Stdout, &devslog.Options{
			SortKeys:           true,
			MaxErrorStackTrace: 10,
			MaxSlicePrintSize:  50,
			HandlerOptions:     opts,
		})
		return traceid.LogHandler(base)
	}

	base := slog.NewJSONHandler(os.Stdout, opts)
	return traceid.LogHandler(base)
}

func isDebugHeaderSet(r *http.Request) bool {
	return r.Header.Get("Debug") == "reveal-body-logs"
}
