package fake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
	"k8s.io/apimachinery/pkg/apimachinery"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/cache"
)

type fakeTPR struct {
	Store cache.Store
}

func (f *fakeTPR) HandlePost(req *http.Request) (*http.Response, error) {
	var tpr *ext.PrsnExternalResource

	data, _ := ioutil.ReadAll(req.Body)
	json.Unmarshal(data, &tpr)
	f.Store.Add(tpr)
	return &http.Response{StatusCode: http.StatusCreated}, nil
}

func objBody(object interface{}) io.ReadCloser {
	output, err := json.MarshalIndent(object, "", "")
	if err != nil {
		panic(err)
	}
	return ioutil.NopCloser(bytes.NewReader([]byte(output)))
}

func (f *fakeTPR) HandleGet(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)

	pathElems := strings.Split(req.URL.Path, "/")
	var items []ext.PrsnExternalResource

	if len(pathElems) == 4 {
		rsc := pathElems[3]
		items = f.resources(rsc)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body: objBody(ext.PrsnExternalResourceList{
			Items: items,
		}),
	}, nil
}

func (f *fakeTPR) resources(rsc string) []ext.PrsnExternalResource {
	r := f.Store.List()

	kind := kindFromElem(rsc)
	retval := []ext.PrsnExternalResource{}
	for _, rr := range r {
		obj := rr.(*ext.PrsnExternalResource)
		if obj.Kind == kind {
			retval = append(retval, *obj)
		}
	}
	return retval
}

// HandleRequest is HTTP API handler for our fake client
func (f *fakeTPR) HandleRequest(req *http.Request) (*http.Response, error) {
	switch m := req.Method; {
	case m == http.MethodPost:
		return f.HandlePost(req)
	case m == http.MethodGet:
		return f.HandleGet(req)
	default:
		return nil, fmt.Errorf("unexpected request: %#v\n%#v", req.URL, req)
	}
}

var manager *registered.APIRegistrationManager

// TPRClient returns fake REST client to be used in TPR unit tests.
func TPRClient(objects ...runtime.Object) *fake.RESTClient {
	var schemeGroupVersion = schema.GroupVersion{Group: "prsn.io", Version: "v1"}
	var groupmeta = apimachinery.GroupMeta{
		GroupVersion: schemeGroupVersion,
	}

	var legacyGroupVersion = schema.GroupVersion{Group: "", Version: "v1"}
	var legacyGroupMeta = apimachinery.GroupMeta{
		GroupVersion: legacyGroupVersion,
	}

	m, _ := registered.NewAPIRegistrationManager("")

	m.AddThirdPartyAPIGroupVersions(schemeGroupVersion)
	m.RegisterGroup(groupmeta)

	// very ugly but works
	m.RegisterGroup(legacyGroupMeta)
	m.RegisterVersions([]schema.GroupVersion{legacyGroupVersion})

	f := &fakeTPR{
		Store: objectStore(objects),
	}

	return &fake.RESTClient{
		GroupName:            "prsn.io",
		NegotiatedSerializer: serializer.DirectCodecFactory{CodecFactory: api.Codecs},
		Client:               fake.CreateHTTPClient(f.HandleRequest),
		APIRegistry:          m,
	}
}

func objectStore(objects []runtime.Object) cache.Store {
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)
	for _, obj := range objects {
		if err := store.Add(obj); err != nil {
			panic(err)
		}
	}
	return store
}

func kindFromElem(e string) string {
	switch e {
	case "mysqls":
		return "Mysql"
	case "mongos":
		return "Mongo"
	case "redises":
		return "Redis"
	case "cassandras":
		return "Cassandra"
	case "zookeepers":
		return "Zookeeper"
	case "kafkas":
		return "Kafka"
	default:
		return "None"
	}
}
