package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var Deploys = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "eo_deploys_total",
		Help: "Deploy requests received from clients.",
	},
	[]string{"status"},
)

func init() {
	prometheus.MustRegister(Deploys)
	Deploys.With(prometheus.Labels{"status": "failed"}).Inc()
	Deploys.With(prometheus.Labels{"status": "succeeded"}).Inc()
}
