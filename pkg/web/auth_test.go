package web

import (
	"testing"
)

func TestAuthToken(t *testing.T) {
	auth := &AuthClient{Token: "asd"}

	if !auth.Authenticate("asd") {
		t.Errorf("Token authentication failed")
	}
}
