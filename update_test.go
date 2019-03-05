package bongoz

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	conn := getConnection()
	collection := conn.Collection("pages")
	defer conn.Session.Close()

	Convey("Update", t, func() {
		endpoint := NewEndpoint("/pages", conn, "pages")
		Convey("basic", func() {
			endpoint.Factory = Factory

			router := getHandler(endpoint)
			w := httptest.NewRecorder()

			obj := &Page{
				Content:  "Foo",
				IntValue: 5,
			}

			err := collection.Save(obj)
			So(err, ShouldEqual, nil)

			updated := map[string]string{
				"Content": "bar",
			}

			marshaled, err := json.Marshal(updated)

			So(err, ShouldEqual, nil)

			reader := strings.NewReader(string(marshaled))
			req, _ := http.NewRequest("POST", strings.Join([]string{"/api/pages", obj.Id.Hex()}, "/"), reader)
			router.ServeHTTP(w, req)

			response := map[string]interface{}{}

			So(w.Code, ShouldEqual, 200)

			err = json.Unmarshal(w.Body.Bytes(), &response)

			So(err, ShouldEqual, nil)

			So(response["Content"], ShouldEqual, "bar")
			So(response["IntValue"], ShouldEqual, 5.0)
		})
		Convey("validation errors", func() {
			endpoint.Factory = ValidFactory

			router := getHandler(endpoint)
			w := httptest.NewRecorder()

			obj := &validatedModel{
				Content: "Biff",
			}

			err := collection.Save(obj)

			So(err, ShouldEqual, nil)

			update := map[string]string{
				"Content": "",
			}

			marshaled, err := json.Marshal(update)

			So(err, ShouldEqual, nil)

			reader := strings.NewReader(string(marshaled))
			req, _ := http.NewRequest("POST", strings.Join([]string{"/api/pages", obj.Id.Hex()}, "/"), reader)
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(w.Body.String(), ShouldEqual, "{\"Error\":\"Validation failed. (Content is required)\"}")
		})

		Reset(func() {
			conn.Session.DB("bongoz").DropDatabase()

		})
	})
}
