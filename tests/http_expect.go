package tests

import (
	"fmt"
	"github.com/gavv/httpexpect/v2"
	"github.com/soffa-projects/go-micro/micro"
	"github.com/soffa-projects/go-micro/schema"
	"github.com/soffa-projects/go-micro/util/digest"
	"github.com/soffa-projects/go-micro/util/h"
	"net/http"
	"net/http/httptest"
	"testing"
)

// HttpExpect is for http testing
type HttpExpect struct {
	t        *testing.T
	http     *httpexpect.Expect
	Teardown func()
}

type HttpTestResult struct {
	t      *testing.T
	result *httpexpect.Response
}

type HttpRequest struct {
	t             *testing.T
	method        string
	form          bool
	path          string
	body          interface{}
	internal      *httpexpect.Expect
	params        any
	authorization string
	headers       map[string]string
}

type ValueExpect struct {
	value *httpexpect.Value
}

type ArrayExpect struct {
	value *httpexpect.Array
}

type NumberExect struct {
	value *httpexpect.Number
}

type StringExpect struct {
	value *httpexpect.String
}

type BooleanExpect struct {
	value *httpexpect.Boolean
}

type ObjectExpect struct {
	value *httpexpect.Object
}

func HttpTest(t *testing.T, handler http.Handler, teardown func()) HttpExpect {
	server := httptest.NewServer(handler)
	return HttpExpect{
		t: t,
		http: httpexpect.WithConfig(httpexpect.Config{
			TestName: t.Name(),
			BaseURL:  server.URL,
			Reporter: httpexpect.NewFatalReporter(t),
			Printers: []httpexpect.Printer{
				httpexpect.NewCompactPrinter(t),
			},
		}),
		Teardown: func() {
			server.Close()
			teardown()
		},
	}
}

type CrudTestConfig struct {
	Bearer      string
	IdPrefix    string
	GetListPath string
	CreateInput h.Map
	UpdateInput h.Map
	SearchInput *schema.FilterInput
}

func (f *HttpExpect) CRUD(path string, config CrudTestConfig) {
	// test list
	f.GET(path).
		BearerAuth(config.Bearer).Expect().IsOK().
		JSON().Path(config.GetListPath).Array().IsEmpty()

	modelId := f.POST(path, config.CreateInput).
		BearerAuth(config.Bearer).Expect().IsOK().
		JSON().Path("$.id").String().HasPrefix(config.IdPrefix).Raw()

	if modelId == "" {
		f.t.Fatalf("failed to create model %s", path)
	}

	f.GET(path).
		BearerAuth(config.Bearer).Expect().IsOK().
		JSON().Path(config.GetListPath).Array().NotEmpty()

	idPath := fmt.Sprintf("%s/%s", path, modelId)

	if len(config.UpdateInput) > 0 {
		result := f.PATCH(idPath, config.UpdateInput).
			BearerAuth(config.Bearer).Expect().IsOK()

		result.JSON().Path("$.id").String().IsEqual(modelId)
		for k, v := range config.UpdateInput {
			result.JSON().Path(fmt.Sprintf("$.%s", k)).String().IsEqual(fmt.Sprintf("%v", v))
		}
	}

	if config.SearchInput != nil {
		f.POST(fmt.Sprintf("%s/search", path), config.SearchInput).
			BearerAuth(config.Bearer).Expect().IsOK().
			JSON().Path(config.GetListPath).Array().NotEmpty()
	}
	f.DELETE(idPath).
		BearerAuth(config.Bearer).Expect().IsOK().JSON().Path("$.id").String().IsEqual(modelId)

	f.GET(path).
		BearerAuth(config.Bearer).Expect().IsOK().
		JSON().Path(config.GetListPath).Array().IsEmpty()
}

func (f *HttpExpect) GET(path string) *HttpRequest {
	return f.request(http.MethodGet, path)
}

func (f *HttpExpect) POST(path string, body ...interface{}) *HttpRequest {
	return f.request(http.MethodPost, path, body...)
}
func (f *HttpExpect) POSTForm(path string, body ...interface{}) *HttpRequest {
	r := &HttpRequest{
		t:        f.t,
		internal: f.http,
		method:   http.MethodPost,
		form:     true,
		path:     path,
		headers:  map[string]string{},
	}
	if body != nil {
		r.body = body[0]
	}
	return r
}

func (f *HttpExpect) PUT(path string, body ...interface{}) *HttpRequest {
	return f.request(http.MethodPut, path, body...)
}

func (f *HttpExpect) DELETE(path string, body ...interface{}) *HttpRequest {
	return f.request(http.MethodDelete, path, body...)
}

func (f *HttpExpect) PATCH(path string, body ...interface{}) *HttpRequest {
	return f.request(http.MethodPatch, path, body...)
}

func (f *HttpExpect) request(method string, path string, body ...interface{}) *HttpRequest {
	r := &HttpRequest{
		t:        f.t,
		internal: f.http,
		method:   method,
		path:     path,
		headers:  map[string]string{},
	}

	if body != nil {
		r.body = body[0]
	}
	return r
}

func (r *HttpRequest) Params(params any) *HttpRequest {
	r.params = params
	return r
}

func (r *HttpRequest) Expect() *HttpTestResult {
	req := r.internal.Request(r.method, r.path)
	if r.body != nil {
		if r.form {
			req = req.WithForm(r.body)
		} else {
			req = req.WithJSON(r.body)
		}
	}
	if r.params != nil {
		req = req.WithQueryObject(r.params)
	}
	if r.authorization != "" {
		req = req.WithHeader("Authorization", r.authorization)
	}
	if r.headers != nil && len(r.headers) > 0 {
		for k, v := range r.headers {
			req = req.WithHeader(k, v)
		}
	}
	return &HttpTestResult{
		t:      r.t,
		result: req.Expect(),
	}
}

func (r *HttpRequest) TenantId(value string) *HttpRequest {
	r.headers[micro.TenantIdHttpHeader] = value
	return r
}

func (r *HttpRequest) Host(value string) *HttpRequest {
	r.headers["Host"] = value
	return r
}

func (r *HttpRequest) Header(name string, value string) *HttpRequest {
	r.headers[name] = value
	return r
}

func (r *HttpRequest) BearerAuth(token string) *HttpRequest {
	if !h.IsStrEmpty(token) {
		r.authorization = "Bearer " + token
	}
	return r
}

func (r *HttpRequest) BasicAuth(username string, password string) *HttpRequest {
	if !h.IsStrEmpty(username) && !h.IsStrEmpty(password) {
		r.authorization = "Basic " + digest.EncodeToBase64String(fmt.Sprintf("%s:%s", username, password))
	}
	return r
}

func (r *HttpTestResult) IsOK() *HttpTestResult {
	r.result.Status(http.StatusOK)
	return r
}
func (r *HttpTestResult) IsCreated() *HttpTestResult {
	r.result.Status(http.StatusCreated)
	return r
}

func (r *HttpTestResult) IsConflict() *HttpTestResult {
	r.result.Status(http.StatusConflict)
	return r
}

func (r *HttpTestResult) IsBadRequest() *HttpTestResult {
	r.result.Status(http.StatusBadRequest)
	return r
}

func (r *HttpTestResult) IsUnauthorized() *HttpTestResult {
	r.result.Status(http.StatusUnauthorized)
	return r
}

func (r *HttpTestResult) IsForbidden() *HttpTestResult {
	r.result.Status(http.StatusForbidden)
	return r
}

func (r *HttpTestResult) Status(status int) *HttpTestResult {
	r.result.Status(status)
	return r
}

func (r *HttpTestResult) JSON() *ValueExpect {
	return &ValueExpect{
		value: r.result.JSON(),
	}
}

func (v *ValueExpect) Object() *ObjectExpect {
	return &ObjectExpect{
		value: v.value.Object(),
	}
}

func (v *ValueExpect) Array() *ArrayExpect {
	return &ArrayExpect{
		value: v.value.Array(),
	}
}

func (v *ValueExpect) Path(path string) *ValueExpect {
	return &ValueExpect{
		value: v.value.Path(path),
	}
}

func (v *ObjectExpect) Path(path string) *ValueExpect {
	return &ValueExpect{
		value: v.value.Path(path),
	}
}

func (v *ObjectExpect) NotContainsKey(path string) *ObjectExpect {
	v.value.NotContainsKey(path)
	return v
}

func (v *ObjectExpect) ContainsKey(path string) *ObjectExpect {
	v.value.ContainsKey(path)
	return v
}

func (v *ValueExpect) IsString() *ValueExpect {
	return &ValueExpect{
		value: v.value.IsString(),
	}
}
func (v *ValueExpect) String() *StringExpect {
	return &StringExpect{
		value: v.value.String(),
	}
}

func (v *ValueExpect) Boolean() *BooleanExpect {
	return &BooleanExpect{
		value: v.value.Boolean(),
	}
}

func (v *ValueExpect) Number() *NumberExect {
	return &NumberExect{
		value: v.value.Number(),
	}
}

func (v *StringExpect) Raw() string {
	return v.value.Raw()
}

func (v *BooleanExpect) IsTrue() *BooleanExpect {
	v.value.IsTrue()
	return v
}

func (v *BooleanExpect) IsFalse() *BooleanExpect {
	v.value.IsFalse()
	return v
}

func (v *ArrayExpect) IsEmpty() *ArrayExpect {
	v.value.IsEmpty()
	return v
}

func (v *ArrayExpect) NotEmpty() *ArrayExpect {
	v.value.NotEmpty()
	return v
}

func (v *ArrayExpect) Length() *NumberExect {
	return &NumberExect{value: v.value.Length()}
}

func (v *NumberExect) IsEqual(value interface{}) *NumberExect {
	v.value.IsEqual(value)
	return v
}

func (v *StringExpect) IsEqual(value string) *StringExpect {
	v.value.IsEqual(value)
	return v
}

func (v *StringExpect) HasPrefix(value string) *StringExpect {
	v.value.HasPrefix(value)
	return v
}

func (v *StringExpect) HasSuffix(value string) *StringExpect {
	v.value.HasSuffix(value)
	return v
}

func (v *StringExpect) NotEqual(value string) *StringExpect {
	v.value.NotEqual(value)
	return v
}

func (v *StringExpect) NotEmpty() *StringExpect {
	v.value.NotEmpty()
	return v
}

func (v *StringExpect) IsEmpty() *StringExpect {
	v.value.IsEmpty()
	return v
}
