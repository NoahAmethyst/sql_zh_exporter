package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"sql_zh_exporter/config"
	"sql_zh_exporter/util/log"

	"net/http"
)

func init() {
	prometheus.MustRegister(version.NewCollector("sql_exporter"))
}

func main() {

	var (
		configFile    = flag.String("config.file", "./config.yml", "SQL Exporter config file")
		jobConfigFile = flag.String("job.config.file", "./job_config.yml", "SQL Exporter job config file")
	)

	flag.Parse()

	log.Init(*configFile)

	log.Info().Fields(map[string]interface{}{
		"message":       "Starting sql_exporter",
		"version_info":  version.Info(),
		"build_context": version.BuildContext(),
	}).Send()

	exporter, err := NewExporter(*jobConfigFile)
	if err != nil {
		log.Panic().Fields(map[string]interface{}{
			"action": "starting exporter",
			"error":  err,
		}).Send()
	}
	prometheus.MustRegister(exporter)

	// setup and start webserver
	metricPath := "/metrics"
	http.Handle(metricPath, promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "OK", http.StatusOK) })
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fmt.Sprintf(`<html>
		<head><title>SQL Exporter</title></head>
		<body>
		<h1>SQL Exporter</h1>
		<p><a href="  %s ">Metrics</a></p>
		</body>
		</html>
		`, metricPath)))
	})

	log.Info().Fields(map[string]interface{}{
		"action":        "start server",
		"listenAddress": config.GetConfig(ServerPort, *configFile),
	}).Send()

	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.GetConfig(ServerPort, *configFile)), nil); err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "starting HTTP server",
			"error":  err,
		}).Send()
	}
}
