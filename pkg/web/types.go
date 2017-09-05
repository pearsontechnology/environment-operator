package web

import "github.com/pearsontechnology/environment-operator/pkg/bitesize"

// DeployRequest represents POST request body to perform deployments.
//  * Name of the service to update
//  * Application image part (full construct from util.DockerImage )
//  * Version application version
type DeployRequest struct {
	Name        string `json:"name"`
	Application string `json:"application,omitempty"`
	Version     string
}

type StatusResponse struct {
	EnvironmentName string          `json:"environment"`
	Namespace       string          `json:"namespace"`
	Services        []StatusService `json:"services"`
}

type StatusService struct {
	Name       string         `json:"name"`
	Version    string         `json:"version,omitempty"`
	URL        string         `json:"external_url,omitempty"`
	DeployedAt string         `json:"deployed_at,omitempty"`
	Replicas   StatusReplicas `json:"replicas,omitempty"`
	Status     string         `json:"status,omitempty"`
}

type StatusPods struct {
	Pods []bitesize.Pod `json:"pods,omitempty"`
}

type StatusReplicas struct {
	Available int `json:"available"`
	UpToDate  int `json:"up_to_date"`
	Desired   int `json:"desired"`
}
