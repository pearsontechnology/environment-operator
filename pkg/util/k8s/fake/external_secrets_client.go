package fake

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type fakeExternalSecret struct {
	Store cache.Store
}

func (f *fakeExternalSecret) HandlePost(req *http.Request) (*http.Response, error) {
	var es *ext.ExternalSecret

	data, _ := ioutil.ReadAll(req.Body)
	_ = json.Unmarshal(data, &es)
	_ = f.Store.Add(es)
	return &http.Response{StatusCode: http.StatusCreated}, nil
}

func (f *fakeExternalSecret) HandleGet(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)

	pathElems := strings.Split(req.URL.Path, "/")
	var items []ext.ExternalSecret

	if len(pathElems) == 4 {
		rsc := pathElems[3]
		items = f.resources(rsc)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body: objBody(ext.ExternalSecretList{
			Items: items,
		}),
	}, nil
}

func (f *fakeExternalSecret) resources(rsc string) []ext.ExternalSecret {
	r := f.Store.List()

	kind := kindFromElem(rsc)
	retval := []ext.ExternalSecret{}
	for _, rr := range r {
		obj := rr.(*ext.ExternalSecret)
		if obj.Kind == kind {
			retval = append(retval, *obj)
		}
	}
	return retval
}

func (f *fakeExternalSecret) HandleRequest(req *http.Request) (*http.Response, error) {
	switch m := req.Method; {
	case m == http.MethodPost:
		return f.HandlePost(req)
	case m == http.MethodGet:
		return f.HandleGet(req)
	default:
		return nil, fmt.Errorf("unexpected request: %#v\n%#v", req.URL, req)
	}
}