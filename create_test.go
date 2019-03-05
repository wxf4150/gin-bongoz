package bongoz

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


func TestCreate(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	Convey("POST", t, func() {
		endpoint := NewEndpoint("/pages", conn, "pages")

		Convey("Basic create", func() {
			endpoint.Factory = Factory

//			router := endpoint.GetRouter()
			handler:=getHandler(endpoint)
			w := httptest.NewRecorder()

			reader := strings.NewReader(`{"content":"foo","idValue":null, "_id":"540e05189b2212ee6b1f44d3"}`)
			req, _ := http.NewRequest("POST", "/api/pages", reader)
			handler.ServeHTTP(w, req)

			response := map[string]interface{}{}
			So(w.Code, ShouldEqual, 201)
			err := json.Unmarshal(w.Body.Bytes(), &response)
			t.Log(string(w.Body.Bytes()))
			So(err, ShouldEqual, nil)

			So(response["Content"], ShouldEqual, "foo")
			So(response["_id"], ShouldEqual, "540e05189b2212ee6b1f44d3")
		})


		Convey("Create with validation errors", func() {
			endpoint.Factory = ValidFactory

//			router := endpoint.GetRouter()
			handler:=getHandler(endpoint)
			w := httptest.NewRecorder()

			obj1 := map[string]string{
				"Content": "",
			}

			marshaled, err := json.Marshal(obj1)

			So(err, ShouldEqual, nil)

			reader := strings.NewReader(string(marshaled))
			req, _ := http.NewRequest("POST", "/api/pages", reader)
			handler.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
		})
		Reset(func() {
			conn.Session.DB("bongoz").DropDatabase()
		})
	})
}
