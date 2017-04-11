package fake

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/testing"

	ext "github.com/pearsontechnology/environment-operator/pkg/k8_extensions"
)

// TPRStore acts as a in-memory "cluster" database for TPR tests
var TPRStore testing.ObjectTracker

// TPRClient returns fake REST client to be used in TPR unit tests.
func TPRClient(objects ...runtime.Object) rest.Interface {
	// var config *rest.Config
	//
	// groupversion := schema.GroupVersion{
	// 	Group:   "prsn.io",
	// 	Version: "v1",
	// }
	//
	// config.GroupVersion = &groupversion
	// config.APIPath = "/apis"
	// config.ContentType = runtime.ContentTypeJSON
	// config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}
	//
	// schemeBuilder := runtime.NewSchemeBuilder(
	// 	func(scheme *runtime.Scheme) error {
	// 		scheme.AddKnownTypes(
	// 			groupversion,
	// 			&ext.PrsnExternalResource{},
	// 			&ext.PrsnExternalResourceList{},
	// 		)
	// 		return nil
	// 	})
	// metav1.AddToGroupVersion(runtime.Scheme, groupversion)
	// schemeBuilder.AddToScheme(runtime.Scheme)

	TPRStore = objectStore(objects)
	return &fake.RESTClient{
		GroupName:            "prsn.io",
		NegotiatedSerializer: serializer.DirectCodecFactory{CodecFactory: api.Codecs},
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/mysql" && m == http.MethodGet:
				var tpr *ext.PrsnExternalResource

				data, _ := ioutil.ReadAll(req.Body)
				json.Unmarshal(data, &tpr)
				TPRStore.Add(tpr)
				return &http.Response{StatusCode: http.StatusCreated}, nil
			default:
				return nil, fmt.Errorf("unexpected request: %#v\n%#v", req.URL, req)
			}
		}),
	}
}

func objectStore(objects []runtime.Object) testing.ObjectTracker {
	o := testing.NewObjectTracker(api.Scheme, api.Codecs.UniversalDecoder())

	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}
	return o
}
