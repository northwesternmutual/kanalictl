// Copyright (c) 2017 Northwestern Mutual.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package utils

import (
	"bytes"
	"net/http"

	"io/ioutil"
)

// HTTPClient is an interface used to perform HTTP GET operations
type HTTPClient interface {
	Get(string) (*http.Response, error)
}

// MockClient is a factory used to create a mock http client for testing
type MockClient struct{}

// Get performs an HTTP GET mock operation
func (m *MockClient) Get(url string) (resp *http.Response, err error) {
	stringBody := `{"kind":"ApiProxyList","items":[{"apiVersion":"kanali.io/v1","kind":"ApiProxy","metadata":{"name":"example-one","namespace":"application","selfLink":"/apis/kanali.io/v1/namespaces/application/apiproxies/example-one","uid":"76187e2d-551d-11e7-a4a3-080027d0fbf8","resourceVersion":"8482","creationTimestamp":"2017-06-19T18:31:08Z","annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"kanali.io/v1\",\"kind\":\"ApiProxy\",\"metadata\":{\"annotations\":{},\"name\":\"example-one\",\"namespace\":\"application\"},\"spec\":{\"path\":\"/api/v1/example-one\",\"service\":{\"name\":\"example-one\",\"port\":8080},\"target\":\"/\"}}\n"}},"spec":{"path":"/api/v1/example-one","service":{"name":"example-one","port":8080},"target":"/"}}],"metadata":{"selfLink":"/apis/kanali.io/v1/apiproxies","resourceVersion":"21991"},"apiVersion":"kanali.io/v1"}`
	if url == "/apis/kanali.io/v1/apikeybindings" {
		stringBody = `{"kind":"ApiKeyBindingList","items":[{"apiVersion":"kanali.io/v1","kind":"ApiKeyBinding","metadata":{"name":"example-eight","namespace":"application","selfLink":"/apis/kanali.io/v1/namespaces/application/apikeybindings/example-eight","uid":"169e2fa0-555f-11e7-a4a3-080027d0fbf8","resourceVersion":"26284","creationTimestamp":"2017-06-20T02:20:54Z","annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"kanali.io/v1\",\"kind\":\"ApiKeyBinding\",\"metadata\":{\"annotations\":{},\"name\":\"example-eight\",\"namespace\":\"application\"},\"spec\":{\"keys\":[{\"name\":\"example-eight-apikey\",\"subpaths\":[{\"path\":\"/foo\",\"rule\":{\"granular\":{\"verbs\":[\"GET\"]}}}]}],\"proxy\":\"example-eight\"}}\n"}},"spec":{"keys":[{"name":"example-eight-apikey","subpaths":[{"path":"/foo","rule":{"granular":{"verbs":["GET"]}}}]}],"proxy":"example-eight"}}],"metadata":{"selfLink":"/apis/kanali.io/v1/apikeybindings","resourceVersion":"26319"},"apiVersion":"kanali.io/v1"}`
	}
	resp = &http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer([]byte(stringBody))),
	}
	return resp, nil
}
