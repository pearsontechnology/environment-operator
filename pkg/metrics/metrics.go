package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var Deploys = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "deploys_total",
		Help: "Number of requests to deploy a service received by EO.",
	},
	[]string{"status"},
)

func init() {
	prometheus.MustRegister(Deploys)
}
