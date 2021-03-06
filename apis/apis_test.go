package apis

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/content"
	"github.com/rogersole/payments-basic-api/app"
	"github.com/rogersole/payments-basic-api/dtos"
	"github.com/rogersole/payments-basic-api/testdata"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type apiTestCase struct {
	tag      string
	method   string
	url      string
	body     string
	status   int
	response string
}

func newRouter() *routing.Router {
	logger := logrus.New()
	logger.Level = logrus.PanicLevel

	router := routing.New()

	router.Use(
		app.Init(logger),
		content.TypeNegotiator(content.JSON),
		app.Transactional(testdata.DB),
	)

	return router
}

func testAPI(router *routing.Router, method, URL, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, URL, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func runAPITests(t *testing.T, router *routing.Router, tests []apiTestCase) {
	for _, test := range tests {
		res := testAPI(router, test.method, test.url, test.body)
		assert.Equal(t, test.status, res.Code, test.tag)
		if test.response != "" {
			assert.JSONEq(t, test.response, res.Body.String(), test.tag)
		}
	}
}

func runAPITestsIgnoringResponseId(t *testing.T, router *routing.Router, tests []apiTestCase) {
	for _, test := range tests {
		res := testAPI(router, test.method, test.url, test.body)
		assert.Equal(t, test.status, res.Code, test.tag)
		if test.response != "" {
			var responsePayment dtos.Payment
			json.Unmarshal(res.Body.Bytes(), &responsePayment)
			responsePayment.Id = uuid.Nil
			responseJSON, _ := json.Marshal(responsePayment)
			var testPayment dtos.Payment
			json.Unmarshal([]byte(test.response), &testPayment)
			testPayment.Id = uuid.Nil
			testResponseJSON, _ := json.Marshal(testPayment)
			assert.JSONEq(t, string(testResponseJSON), string(responseJSON), test.tag)
		}
	}
}
