package metrics

import (
	"github.com/k3s-io/kine/pkg/database/sql"
	"fmt"
	"time"

	"github.com/k3s-io/kine/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	ResultSuccess = "success"
	ResultError   = "error"
)

var (
	SQLTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kine_sql_total",
		Help: "Total number of SQL operations",
	}, []string{"error_code"})

	SQLTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "kine_sql_time_seconds",
		Help: "Length of time per SQL operation",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
			1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30},
	}, []string{"error_code"})

	CompactTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kine_compact_total",
		Help: "Total number of compactions",
	}, []string{"result"})

	InsertErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kine_insert_errors_total",
		Help: "Total number of insert retries due to unique constraint violations",
	}, []string{"retriable"})
)

var (
	// SlowSQLThreshold is a duration which SQL executed longer than will be logged.
	// This can be directly modified to override the default value when kine is used as a library.
	SlowSQLThreshold        = time.Second
	SlowSQLWarningThreshold = 5 * time.Second
)

func ObserveSQL(start time.Time, errCode string, sql util.Stripped, args ...interface{}) {
	SQLTotal.WithLabelValues(errCode).Inc()
	duration := time.Since(start)
	SQLTime.WithLabelValues(errCode).Observe(duration.Seconds())
	if SlowSQLThreshold > 0 && duration >= SlowSQLThreshold {
		instrumentedLogger := logrus.WithField("duration", duration)

		if logrus.GetLevel() == logrus.TraceLevel {
			instrumentedLogger.WithField("args", args)
		}

		if duration < SlowSQLWarningThreshold {
			instrumentedLogger.Infof("Slow SQL (started: %v) (total time: %v): %s", start, duration, sql)
		} else {
			instrumentedLogger.Warnf("Slow SQL (started: %v) (total time: %v): %s", start, duration, sql)
		}
	}
}

func WriteDBStats(id int, name string, stats sql.DBStats) {
	instrumentedLogger := logrus.WithField("dbstatsName", name)
	s := fmt.Sprintf("maxOpenConnections: %d, openConections: %d, inUse: %d, idle: %d, waitCount: %d, waitDuration: %s, maxIdleClosed: %d, maxIdleTimeClosed %d, maxLifetimeClosed: %d",
		stats.MaxOpenConnections,
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
		stats.MaxIdleClosed,
		stats.MaxIdleTimeClosed,
		stats.MaxLifetimeClosed,
	)
	instrumentedLogger.Infof("ID: %d, DB stats: %s", id, s)
}
