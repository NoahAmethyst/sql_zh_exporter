package main

import (
	"fmt"
	"sql_zh_exporter/util/log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // register the PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cenkalti/backoff/v4"
	//达梦数据库驱动
	_ "github.com/team-ide/go-driver/db_dm"
	//人大金仓数据库驱动
	_ "github.com/team-ide/go-driver/db_kingbase_v8r6"
	//神通数据库驱动
	//_ "github.com/team-ide/go-driver/db_shentong"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	// MetricNameRE matches any invalid metric name
	// characters, see github.com/prometheus/common/model.MetricNameRE
	MetricNameRE = regexp.MustCompile("[^a-zA-Z0-9_:]+")
)

// Init will initialize the metric descriptors
func (j *Job) Init(queries map[string]string) error {
	// register each query as an metric
	for _, q := range j.Queries {
		if q == nil {
			log.Warn().Fields(map[string]interface{}{
				"query": q.Name,
				"warn":  "Skipping invalid query",
			}).Send()

			continue
		}
		q.jobName = j.Name
		if q.Query == "" && q.QueryRef != "" {
			if qry, found := queries[q.QueryRef]; found {
				q.Query = qry
			}
		}
		if q.Query == "" {
			log.Warn().Fields(map[string]interface{}{
				"query": q.Name,
				"warn":  "Skipping empty query",
			}).Send()

			continue
		}
		if q.metrics == nil {
			// we have no way of knowing how many metrics will be returned by the
			// queries, so we just assume that each query returns at least one metric.
			// after the each round of collection this will be resized as necessary.
			q.metrics = make(map[*connection][]prometheus.Metric, len(j.Queries))
		}
		// try to satisfy prometheus naming restrictions
		name := MetricNameRE.ReplaceAllString(fmt.Sprintf("sql_%s_%s", q.jobName, q.Name), "")
		//name := MetricNameRE.ReplaceAllString(fmt.Sprintf("%sdb_%s", q.jobName, q.Name), "")
		//name := strings.ToLower(fmt.Sprintf("%s_%s", q.jobName, q.Name))
		//name = MetricNameRE.ReplaceAllString(name, "")
		help := q.Help
		// prepare a new metrics descriptor
		//
		// the tricky part here is that the *order* of labels has to match the
		// order of label values supplied to NewConstMetric later
		q.desc = prometheus.NewDesc(
			name,
			help,
			append(q.Labels, "driver", "host", "database", "user", "col"),
			prometheus.Labels{
				"sql_job": j.Name,
			},
		)
	}
	j.updateConnections()
	return nil
}

func (j *Job) updateConnections() {
	// if there are no connection URLs for this job it can't be run
	if j.Connections == nil {
		log.Error().Fields(map[string]interface{}{
			"error": "No connections for job",
			"job":   j.Name,
		}).Send()
		return
	}
	// make space for the connection objects
	if j.conns == nil {
		j.conns = make([]*connection, 0, len(j.Connections))
	}
	// parse the connection URLs and create an connection object for each
	if len(j.conns) < len(j.Connections) {
		for _, conn := range j.Connections {
			u, err := url.Parse(conn)
			if err != nil {
				log.Error().Fields(map[string]interface{}{
					"action": "Failed to parse URL",
					"job":    j.Name,
					"url":    conn,
					"error":  err,
				}).Send()
				continue
			}
			user := ""

			if u.User != nil {
				user = u.User.Username()
			}
			// we expose some connection variables as labels, so we need to
			// remember them
			newConn := &connection{
				conn:     nil,
				url:      conn,
				driver:   u.Scheme,
				host:     u.Host,
				database: strings.TrimPrefix(u.Path, "/"),
				user:     user,
			}
			j.conns = append(j.conns, newConn)
		}
	}
}

func (j *Job) ExecutePeriodically() {

	log.Debug().Fields(map[string]interface{}{
		"action": "Starting",
		"job":    j.Name,
	}).Send()

	for {
		j.Run()
		log.Debug().Fields(map[string]interface{}{
			"action": "Sleeping until next run",
			"job":    j.Name,
			"sleep":  j.Interval.String(),
		}).Send()
		time.Sleep(j.Interval)
	}
}

func (j *Job) runOnceConnection(conn *connection, done chan int) {
	updated := 0
	defer func() {
		done <- updated
	}()

	// connect to DB if not connected already
	if err := conn.connect(j); err != nil {
		log.Warn().Fields(map[string]interface{}{
			"action": "Failed to connect",
			"warn":   err,
			"job":    j.Name,
		}).Send()
		j.markFailed(conn)
		return
	}

	for _, q := range j.Queries {
		if q == nil {
			continue
		}
		if q.desc == nil {
			// this may happen if the metric registration failed
			log.Warn().Fields(map[string]interface{}{
				"action": "Skipping query",
				"warn":   "Collector is nil",
				"job":    j.Name,
			}).Send()
			continue
		}
		log.Debug().Fields(map[string]interface{}{
			"action": "Running Query",
			"query":  q.Name,
		}).Send()

		// execute the query on the connection
		if err := q.Run(conn); err != nil {
			log.Warn().Fields(map[string]interface{}{
				"action": "Failed to run query",
				"warn":   err,
				"query":  q.Name,
			}).Send()
			continue
		}
		log.Debug().Fields(map[string]interface{}{
			"query":  q.Name,
			"action": "Query finished",
		}).Send()
		updated++
	}
}

func (j *Job) markFailed(conn *connection) {
	for _, q := range j.Queries {
		failedScrapes.WithLabelValues(conn.driver, conn.host, conn.database, conn.user, q.jobName, q.Name).Set(1.0)
	}
}

// Run the job queries with exponential backoff, implements the cron.Job interface
func (j *Job) Run() {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = j.Interval
	if bo.MaxElapsedTime == 0 {
		bo.MaxElapsedTime = time.Minute
	}
	if err := backoff.Retry(j.runOnce, bo); err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "Failed to run",
			"error":  err,
			"job":    j.Name,
		}).Send()
	}
}

func (j *Job) runOnce() error {
	doneChan := make(chan int, len(j.conns))

	// execute queries for each connection in parallel
	for _, conn := range j.conns {
		go j.runOnceConnection(conn, doneChan)
	}

	// connections now run in parallel, wait for and collect results
	updated := 0
	for range j.conns {
		updated += <-doneChan
	}

	if updated < 1 {
		return fmt.Errorf("zero queries ran")
	}
	return nil
}

func (c *connection) connect(job *Job) error {
	// already connected
	if c.conn != nil {
		return nil
	}
	dsn := c.url
	switch c.driver {
	case "mysql":
		dsn = strings.TrimPrefix(dsn, "mysql://")
	case "clickhouse":
		dsn = "tcp://" + strings.TrimPrefix(dsn, "clickhouse://")
	}

	conn, err := sqlx.Connect(c.driver, dsn)
	if err != nil {
		return err
	}
	// be nice and don't use up too many connections for mere metrics
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	// Disable SetConnMaxLifetime if MSSQL as it is causing issues with the MSSQL driver we are using. See #60
	if c.driver != "sqlserver" {
		conn.SetConnMaxLifetime(job.Interval * 2)
	}

	// execute StartupSQL
	for _, query := range job.StartupSQL {
		log.Debug().Fields(map[string]interface{}{
			"msg":   "StartupSQL",
			"Query": query,
			"job":   job.Name,
		}).Send()

		conn.MustExec(query)
	}

	c.conn = conn
	return nil
}
