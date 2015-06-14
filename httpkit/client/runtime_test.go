package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-swagger/go-swagger/client"
	"github.com/go-swagger/go-swagger/httpkit"
	"github.com/go-swagger/go-swagger/spec"
	"github.com/go-swagger/go-swagger/strfmt"
	"github.com/stretchr/testify/assert"
)

// task This describes a task. Tasks require a content property to be set.
type task struct {

	// Completed
	Completed bool `json:"completed"`

	// Content Task content can contain [GFM](https://help.github.com/articles/github-flavored-markdown/).
	Content string `json:"content"`

	// ID This id property is autogenerated when a task is created.
	ID int64 `json:"id"`
}

func TestRuntime_Canary(t *testing.T) {
	// test that it can make a simple request
	// and get the response for it.
	// defaults all the way down
	result := []task{
		{false, "task 1 content", 1},
		{false, "task 2 content", 2},
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime)
		rw.WriteHeader(http.StatusOK)
		jsongen := json.NewEncoder(rw)
		jsongen.Encode(result)
	}))

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	specDoc, err := spec.Load("../../fixtures/codegen/todolist.simple.yml")
	hu, _ := url.Parse(server.URL)
	specDoc.Spec().Host = hu.Host
	specDoc.Spec().BasePath = "/"
	if assert.NoError(t, err) {
		runtime := New(specDoc)
		res, err := runtime.Submit("getTasks", rwrtr, client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				var result []task
				if err := consumer.Consume(response.Body(), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, errors.New("Generic error")
		}))
		if assert.NoError(t, err) {
			assert.IsType(t, []task{}, res)
			actual := res.([]task)
			assert.EqualValues(t, result, actual)
		}
	}
}
