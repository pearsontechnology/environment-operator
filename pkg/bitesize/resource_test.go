package bitesize

import (
	"os"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	t.Run("configmap parsed correctly", testUmarshalConfigMap)
}

func TestFind(t *testing.T) {
	t.Run("find existing configmap by path", testFindConfigMapByPath)
	t.Run("find non-existing configmap by path", testFindConfigMapByPathNotFound)
	t.Run("find existing job by path", testFindJobByPathNotFound)
	t.Run("find non-existing job by path", testFindJobByPathNotFound)
	t.Run("find existing cronjob by path", testFindCronJobByPath)
	t.Run("find non-existing cronjob by path", testFindCronJobByPathNotFound)
}

func testFindConfigMapByPath(t *testing.T) {
	var svc = Imports{
		Resource{
			Name: "a",
			Path: "k8s/a.properties",
			Type: TypeConfigMap,
		},
		Resource{
			Name: "c",
			Path: "k8s/b.properties",
			Type: TypeCronJob,
		},
		Resource{
			Name: "b",
			Path: "k8s/c.properties",
			Type: TypeJob,
		},
	}

	if s := svc.Find("k8s/a.properties", TypeConfigMap); s == nil {
		t.Errorf("Expected resource, got %v", s)
	}
}

func testFindConfigMapByPathNotFound(t *testing.T) {
	var im = Imports{
		Resource{
			Name: "a",
			Path: "k8s/a.properties",
			Type: TypeCronJob,
		},
		Resource{
			Name: "c",
			Path: "k8s/b.properties",
			Type: TypeCronJob,
		},
		Resource{
			Name: "b",
			Path: "k8s/c.properties",
			Type: TypeJob,
		},
	}

	if s := im.Find("k8s/e.properties", TypeConfigMap); s != nil {
		t.Errorf("Expected nil, got %v", s)
	}
}

func testFindJobByPath(t *testing.T) {
	var svc = Imports{
		Resource{
			Name: "a",
			Path: "k8s/a.properties",
			Type: TypeConfigMap,
		},
		Resource{
			Name: "c",
			Path: "k8s/b.properties",
			Type: TypeCronJob,
		},
		Resource{
			Name: "b",
			Path: "k8s/c.properties",
			Type: TypeJob,
		},
	}

	if s := svc.Find("k8s/c.properties", TypeJob); s == nil {
		t.Errorf("Expected resource, got %v", s)
	}
}

func testFindJobByPathNotFound(t *testing.T) {
	var im = Imports{
		Resource{
			Name: "a",
			Path: "k8s/a.properties",
			Type: TypeConfigMap,
		},
		Resource{
			Name: "c",
			Path: "k8s/b.properties",
			Type: TypeCronJob,
		},
		Resource{
			Name: "b",
			Path: "k8s/c.properties",
			Type: TypeCronJob,
		},
	}

	if s := im.Find("k8s/e.properties", TypeJob); s != nil {
		t.Errorf("Expected nil, got %v", s)
	}
}

func testFindCronJobByPath(t *testing.T) {
	var svc = Imports{
		Resource{
			Name: "a",
			Path: "k8s/a.properties",
			Type: TypeConfigMap,
		},
		Resource{
			Name: "c",
			Path: "k8s/b.properties",
			Type: TypeCronJob,
		},
		Resource{
			Name: "b",
			Path: "k8s/c.properties",
			Type: TypeCronJob,
		},
	}

	if s := svc.Find("k8s/c.properties", TypeCronJob); s == nil {
		t.Errorf("Expected resource, got %v", s)
	}
}

func testFindCronJobByPathNotFound(t *testing.T) {
	var im = Imports{
		Resource{
			Name: "a",
			Path: "k8s/a.properties",
			Type: TypeJob,
		},
		Resource{
			Name: "c",
			Path: "k8s/b.properties",
			Type: TypeJob,
		},
		Resource{
			Name: "b",
			Path: "k8s/c.properties",
			Type: TypeConfigMap,
		},
	}

	if s := im.Find("k8s/e.properties", TypeCronJob); s != nil {
		t.Errorf("Expected nil, got %v", s)
	}
}

func testUmarshalConfigMap(t *testing.T) {
	r := &Resource{
		Name: "application-v1",
		Path: "../../test/assets/k8s/application-v1.bitesize",
		Type: TypeConfigMap,
	}
	pwd, _ := os.Getwd()
	if err := LoadResource(r, "dev", pwd); err != nil {
		t.Errorf("Errors expected nil, got %v", err)
	}
	if r.ConfigMap.GetObjectMeta().GetName() != "application-v1" {
		t.Errorf("Expected application-v1, got %v", r.ConfigMap.GetObjectMeta().GetName())
	}
	if r.ConfigMap.GetNamespace() != "dev" {
		t.Errorf("Expected dev, got %v", r.ConfigMap.GetNamespace())
	}
	if !reflect.DeepEqual(r.ConfigMap.Data, map[string]string{
		"data-2": "value-2",
		"data-3": "value-3",
	}) {
		t.Errorf("Expected %v, got %v", map[string]string{
			"data-2": "value-2",
			"data-3": "value-3",
		}, r.ConfigMap.Data)
	}
}
