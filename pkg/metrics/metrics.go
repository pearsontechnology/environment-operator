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
var ConfigMapDeploys = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "eo_configmap_deploys_total",
		Help: "ConfigMap deploy requests received from clients.",
	},
	[]string{"status"},
)

func init() {
	prometheus.MustRegister(Deploys)
	prometheus.MustRegister(ConfigMapDeploys)
}
