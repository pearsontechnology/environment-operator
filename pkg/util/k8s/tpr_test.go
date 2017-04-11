package k8s

import (
	"testing"
)

func TestTPRGet(t *testing.T) {
	// tprc := fake.TPRClient(
	// 	&ext.PrsnExternalResource{
	// 		TypeMeta: unversioned.TypeMeta{
	// 			Kind: "Mysql",
	// 		},
	// 		ObjectMeta: v1.ObjectMeta{
	// 			Name:      "test",
	// 			Namespace: "test",
	// 		},
	// 	},
	// )
	//
	// client := ThirdPartyResource{
	// 	Interface: tprc,
	// 	Namespace: "test",
	// }
	//
	// _, err := client.Get("test")
	// if err != nil {
	// 	t.Errorf("Got unexpected error: %s", err.Error())
	// }

}

func TestTPRExist(t *testing.T) {
}

func TestTPRApply(t *testing.T) {
}

func TestTPRDestroy(t *testing.T) {
}
