package main

import (
	"flag"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

	"github.com/shayonj/dogstatsd-sift/configuration"

	"github.com/shayonj/dogstatsd-sift/datadog"
	log "github.com/sirupsen/logrus"
)

const origin = "https://app.datadoghq.com"
const logFileName = "dogstatsd_sift_request.log"

func setupLog() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
	} else {
		log.Info("Failed to log to file, using default stdout")
	}
}

func parseConfig() (*configuration.Base, error) {
	configFile := flag.String("config-file", "", "config file location")
	flag.Parse()

	if *configFile != "" {
		c, err := configuration.Parse(*configFile)
		return c, err
	}

	return nil, nil
}

func port(cfg *configuration.Base) string {
	if strconv.Itoa(cfg.Port) != "" {
		return ":" + strconv.Itoa(cfg.Port)
	}

	return ":9000"
}

func logFields(r *http.Request) log.Fields {
	return log.Fields{
		"path":             r.URL.Path,
		"raw_query":        r.URL.RawQuery,
		"host":             r.URL.Host,
		"hostname":         r.Host,
		"content_encoding": r.Header.Get(datadog.ContentEncodingHeader),
		"x_forward_for":    r.Header.Get("X-Forwarded-For"),
		"accept_encoding":  r.Header.Get("Accept-Encoding"),
		"conten_type":      r.Header.Get("Content-Type"),
		"dd_agent_version": r.Header.Get("Dd-Agent-Version"),
		"user_agent":       r.Header.Get("User-Agent"),
	}
}

func main() {
	setupLog()

	cfg, err := parseConfig()
	if err != nil {
		log.Errorf("Failed to parse config %s", err)
		panic(err)
	}
	log.Infof("Configuration parsed successfully")

	remote, err := url.Parse(origin)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc("/", handler(proxy, cfg))

	portOn := port(cfg)
	log.Infof("Starting server. Listening on %s", portOn)

	log.Fatal(http.ListenAndServe(portOn, nil))
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		requestLogger := log.WithFields(logFields(r))
		requestLogger.Info("request received")

		datadog.HandleRequest(r, requestLogger)
		defer r.Body.Close()

		p.ServeHTTP(w, r)
	}
}
