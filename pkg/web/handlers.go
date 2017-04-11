package web

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pearsontechnology/environment-operator/pkg/cluster"
	"github.com/pearsontechnology/environment-operator/pkg/config"
	"github.com/pearsontechnology/environment-operator/pkg/util"
	"github.com/pearsontechnology/environment-operator/pkg/util/k8s"

	"github.com/gorilla/mux"
)

// Router returns mux.Router with all paths served
func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/deploy", postDeploy).Methods("POST")
	r.HandleFunc("/status", getStatus).Methods("GET")

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
	client, err := k8s.ClientForNamespace("sample-app-dev")

	if err != nil {
		log.Errorf("Error creating kubernetes client: %s", err.Error())
	}

	d, err := ParseDeployRequest(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	deployment, err := GetCurrentDeploymentByName(d.Name)
	if err != nil {
		log.Errorf("Error getting deployment %s: %s", d.Name, err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	deployment.Spec.Template.Spec.Containers[0].Image = util.Image(d.Application, d.Version)

	if err = client.Deployment().Update(deployment); err != nil {
		log.Errorf("Error updating deployment %s: %s", d.Name, err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}

func getStatus(w http.ResponseWriter, r *http.Request) {

	client, err := cluster.Client()
	if err != nil {
		log.Errorf("Error getting cluster client: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	cfg := config.Load()
	e, err := client.LoadEnvironment(cfg.Namespace)
	if err != nil {
		log.Errorf("Error getting cluster client: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	s := &StatusResponse{
		EnvironmentName: e.Name,
		Namespace:       e.Namespace,
	}

	for _, svc := range e.Services {
		statusService := StatusService{
			Name:       svc.Name,
			Version:    svc.Version,
			DeployedAt: svc.Status.DeployedAt,
			Replicas: StatusReplicas{
				Available: svc.Status.AvailableReplicas,
				UpToDate:  svc.Status.CurrentReplicas,
				Desired:   svc.Status.DesiredReplicas,
			},
		}
		s.Services = append(s.Services, statusService)
	}
	json.NewEncoder(w).Encode(s)

	// LoadEnvironmentFromCluster
	// services.each
	// ret[name]
	// ret[version]

	// Status will get:
	// list of running pods
	// versions of running pods (label version)
	// pods running vs pods desired
	// events for deployment
	// deployment created_at

	// { name: "asd",
	//   current_replicas: X,
	//   desired_replicas: Y,
	//   created_at: x-x-x-x x:x
	//   instances: [
	//      a : {
	//           version: a,
	//           created_at: x-x-x-x
	//      },
	//      b: {
	//         version: a,
	//         created_at: x-x-x-x
	//      }

}
