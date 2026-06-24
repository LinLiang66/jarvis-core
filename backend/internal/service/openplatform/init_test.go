package openplatform

import "testing"

func TestInitActionRegistry(t *testing.T) {
	if _, ok := GetRegisteredActionMeta(ActionEcho); !ok {
		t.Fatal("echo demo action not registered")
	}
}
