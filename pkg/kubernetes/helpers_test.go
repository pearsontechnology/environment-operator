package kubernetes

// func TestClaimFail(t *testing.T) {
// 	client := fake.NewSimpleClientset()
// 	w := Wrapper{Interface: client}
// 	_, err := w.volumeClaimFromName("test", "test")
// 	if err == nil {
// 		t.Error("Expected error")
// 	} else if err.Error() != "Persistent volume claim test not found" {
// 		t.Errorf("Wrong error message: %s", err.Error())
// 	}
// }
