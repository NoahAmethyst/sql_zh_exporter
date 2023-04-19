package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"sql_zh_exporter/util/log"
)

// Exporter collects SQL metrics. It implements prometheus.Collector.
type Exporter struct {
	jobs          []*Job
	cronScheduler *cron.Cron
}

// NewExporter returns a new SQL Exporter for the provided config.
func NewExporter(configFile string) (*Exporter, error) {

	// read config
	cfg, err := Read(configFile)
	if err != nil {
		return nil, err
	}

	exp := &Exporter{
		jobs:          make([]*Job, 0, len(cfg.Jobs)),
		cronScheduler: cron.New(),
	}

	// dispatch all jobs
	for _, job := range cfg.Jobs {
		if job == nil {
			continue
		}

		if err := job.Init(cfg.Queries); err != nil {
			log.Warn().Fields(
				map[string]interface{}{
					"message": "Skipping job. Failed to initialize",
					"error":   err,
					"job":     job.Name,
				}).Send()

			continue
		}
		exp.jobs = append(exp.jobs, job)
		if job.CronSchedule.schedule != nil {
			exp.cronScheduler.Schedule(job.CronSchedule.schedule, job)
			log.Info().Fields(map[string]interface{}{
				"message":       "Scheduled CRON job",
				"job":           job.Name,
				"cron_schedule": job.CronSchedule.definition,
			}).Send()

		} else {
			go job.ExecutePeriodically()
			log.Info().Fields(map[string]interface{}{
				"message":  "Started periodically execution of job",
				"job":      job.Name,
				"interval": job.Interval.String(),
			}).Send()
		}
	}
	exp.cronScheduler.Start()
	return exp, nil
}

// Describe implements prometheus.Collector
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, job := range e.jobs {
		if job == nil {
			continue
		}
		for _, query := range job.Queries {
			if query == nil {
				continue
			}
			if query.desc == nil {
				log.Error().Fields(map[string]interface{}{
					"error": "Query has no descripto",
					"query": query.Name,
				}).Send()
				continue
			}
			ch <- query.desc
		}
	}
}

// Collect implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for _, job := range e.jobs {
		if job == nil {
			continue
		}
		for _, query := range job.Queries {
			if query == nil {
				continue
			}
			for _, metrics := range query.metrics {
				for _, metric := range metrics {
					ch <- metric
				}
			}
		}
	}
}
