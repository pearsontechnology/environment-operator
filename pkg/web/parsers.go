package web

import (
	"encoding/json"
	"io"
)

// ParseDeployRequest returns DeployRequest struct based on
// HTTP request body
func ParseDeployRequest(body io.Reader) (*DeployRequest, error) {
	var req *DeployRequest

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&req)
	return req, err
}
