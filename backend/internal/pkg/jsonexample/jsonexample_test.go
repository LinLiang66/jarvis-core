package jsonexample

import "testing"

func TestSampleStructWithNilPointerField(t *testing.T) {
	type inner struct {
		Name string `json:"name"`
	}
	type outer struct {
		Inner *inner `json:"inner,omitempty"`
	}
	if Sample(outer{}) == nil {
		t.Fatal("expected sample map")
	}
}

func TestSampleNilPointerResponse(t *testing.T) {
	type resp struct {
		Success bool   `json:"success"`
		TraceId string `json:"traceId"`
	}
	var p *resp
	if BusinessResponse("demo.action", p) == "" {
		t.Fatal("expected business response json")
	}
}
