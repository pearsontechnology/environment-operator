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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/cache"
)

type fakeCRD struct {
	Store cache.Store
}

func (f *fakeCRD) HandlePost(req *http.Request) (*http.Response, error) {
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

func (f *fakeCRD) HandleGet(req *http.Request) (*http.Response, error) {
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

func (f *fakeCRD) resources(rsc string) []ext.PrsnExternalResource {
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
func (f *fakeCRD) HandleRequest(req *http.Request) (*http.Response, error) {
	switch m := req.Method; {
	case m == http.MethodPost:
		return f.HandlePost(req)
	case m == http.MethodGet:
		return f.HandleGet(req)
	default:
		return nil, fmt.Errorf("unexpected request: %#v\n%#v", req.URL, req)
	}
}

var manager *runtime.Scheme

// CRDClient returns fake REST client to be used in TPR unit tests.
func CRDClient(group string, version string, objects ...runtime.Object) *fake.RESTClient {
	var schemeGroupVersion = schema.GroupVersion{Group: group, Version: version}

	f := &fakeCRD{
		Store: objectStore(objects),
	}

	return &fake.RESTClient{
		GroupVersion:         schemeGroupVersion,
		NegotiatedSerializer: serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs},
		Client:               fake.CreateHTTPClient(f.HandleRequest),
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
	case "auroras":
		return "Aurora"
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
	case "postgreses":
		return "Postgres"
	case "snses":
		return "Sns"
	case "neptunes":
		return "Neptune"
	case "msks":
		return "Msk"
	case "docdbs":
		return "Docdb"
	case "cbs":
		return "Cb"
	case "sqses":
		return "Sqs"
	case "s3s":
		return "S3"
	case "helmcharts":
		return "HelmChart"
	case "dynamodbs":
		return "dynamodb"
	case "atlases":
		return "atlas"
	case "eses":
		return "es"
	case "cloudfronts":
		return "cloudfront"
	case "crons":
		return "cron"
	default:
		return "None"
	}
}
