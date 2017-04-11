package web

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
	Pods       []StatusPod    `json:"pods,omitempty"`
}

type StatusReplicas struct {
	Available int `json:"available"`
	UpToDate  int `json:"up_to_date"`
	Desired   int `json:"desired"`
}

type StatusPod struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
