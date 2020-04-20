package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/bitesize"
	"github.com/pearsontechnology/environment-operator/pkg/cluster"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Router returns mux.Router with all paths served
func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/deploy", postDeploy).Methods("POST")
	r.HandleFunc("/status", getStatus).Methods("GET")
	r.HandleFunc("/status/{service}", getServiceStatus).Methods("GET")
	r.HandleFunc("/status/{service}/pods", getPodStatus).Methods("GET")
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		tokens, ok := r.Header["Authorization"]
		if ok && len(tokens) >= 1 {
			token = tokens[0]
			token = strings.TrimPrefix(token, "Bearer ")
		}

		auth, err := NewAuthClient()
		if err != nil {
			log.Error(err)
		}
		if auth.Authenticate(token) {
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	})
}

func postDeploy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	client, err := cluster.Client()

	if err != nil {
		log.Errorf("error creating kubernetes client: %s", err.Error())
	}

	d, err := ParseDeployRequest(r.Body)
	if err != nil {
		log.Errorf("could not parse request body: %s", err.Error())
		http.Error(w, fmt.Sprintf("Bad Request: Unable to parse request body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	service, err := loadServiceFromConfig(d.Name)
	if err != nil {
		log.Errorf("error getting deployment %s: %s", d.Name, err.Error())
		http.Error(w, fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	configmaps, err := loadConfigMapsFromConfig()
	if err != nil {
		log.Errorf("error getting ConfigMaps %s: %s", d.Name, err.Error())
		http.Error(w, fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if service.IsBlueGreenParentDeployment() {
		service, err = loadServiceFromConfig(service.InactiveDeploymentName())
		if err != nil {
			log.Errorf("error getting deployment %s: %s", d.Name, err.Error())
			http.Error(w, fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusBadRequest)
			return
		}
	}

	service.Version = d.Version
	service.Application = d.Application

	if err := client.ApplyService(service, configmaps, config.Env.Namespace); err != nil {
		log.Errorf("error updating deployment %s: %s", d.Name, err.Error())
		http.Error(w, fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusBadRequest)
		metrics.Deploys.With(prometheus.Labels{"status": "failed"}).Inc()
		return
	}
	metrics.Deploys.With(prometheus.Labels{"status": "succeeded"}).Inc()

	status := map[string]string{
		"status": "deploying",
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(status)
	if err != nil {
		log.Error(err)
	}
}

func getStatus(w http.ResponseWriter, r *http.Request) {

	client, err := cluster.Client()
	if err != nil {
		log.Errorf("error getting cluster client: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	e, err := client.LoadEnvironment(config.Env.Namespace)
	if err != nil {
		log.Errorf("error getting cluster client: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	s := &StatusResponse{
		EnvironmentName: e.Name,
		Namespace:       e.Namespace,
	}

	for _, svc := range e.Services {

		if svc.IsBlueGreenParentDeployment() {
			if loadSvc, err := loadServiceFromCluster(svc.InactiveDeploymentName()); err == nil {
				loadSvc.Name = svc.Name
				svc = loadSvc
			}
		}
		status := statusForService(svc)
		s.Services = append(s.Services, status)
	}
	err = json.NewEncoder(w).Encode(s)
	if err != nil {
		log.Error(err)
	}
}

func getPodStatus(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	serviceName := vars["service"]
	w.Header().Set("Content-Type", "application/json")
	client, err := cluster.Client()
	if err != nil {
		log.Errorf("Error getting cluster client: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	pods, err := client.LoadPods(config.Env.Namespace)

	deploySVC, err := loadServiceFromCluster(serviceName)

	if err != nil {
		log.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var deployedPods []bitesize.Pod
	//Only return pods that are part of the deployment/service being requested
	for _, pod := range pods {
		if strings.Contains(pod.Name, deploySVC.Name) {
			deployedPods = append(deployedPods, pod)
		}
	}

	statusPods := StatusPods{
		Pods: deployedPods,
	}

	err = json.NewEncoder(w).Encode(statusPods)
	if err != nil {
		log.Error(err)
	}
}

func getServiceStatus(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	serviceName := vars["service"]

	w.Header().Set("Content-Type", "application/json")

	svc, err := loadServiceFromCluster(serviceName)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if svc.IsBlueGreenParentDeployment() {
		if loadSvc, err := loadServiceFromCluster(svc.InactiveDeploymentName()); err == nil {
			loadSvc.Name = svc.Name
			svc = loadSvc
		}
	}
	status := statusForService(svc)
	err = json.NewEncoder(w).Encode(status)
	if err != nil {
		log.Error(err)
	}
}

func statusForService(svc bitesize.Service) StatusService {
	status := "red"
	if svc.Status.AvailableReplicas == svc.Status.DesiredReplicas {
		status = "orange"
	}

	if svc.Status.AvailableReplicas == svc.Status.DesiredReplicas &&
		svc.Status.DesiredReplicas == svc.Status.CurrentReplicas {
		status = "green"
	}

	return StatusService{
		Name:       svc.Name,
		Version:    svc.Version,
		DeployedAt: svc.Status.DeployedAt,
		Status:     status,
		Replicas: StatusReplicas{
			Available: svc.Status.AvailableReplicas,
			UpToDate:  svc.Status.CurrentReplicas,
			Desired:   svc.Status.DesiredReplicas,
		},
	}
}
