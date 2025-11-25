package appinspect

import (
	"bytes"
	"io"
	"net/http"
)

type MockClient struct {
	ResponseBody string
	StatusCode   int
	Err          error
}

func (m *MockClient) CheckValidationStatus(_ string) (*http.Response, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	resp := &http.Response{
		StatusCode: m.StatusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.ResponseBody)),
	}
	return resp, nil
}
