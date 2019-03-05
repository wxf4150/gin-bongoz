package bongoz

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

)

func TestReadOne(t *testing.T) {
	conn := getConnection()
	collection := conn.Collection("pages")
	defer conn.Session.Close()

	Convey("ReadOne", t, func() {
		endpoint := NewEndpoint("/pages", conn, "pages")
		Convey("basic", func() {
			endpoint.Factory = Factory
			//			router := endpoint.GetRouter()
			handler:=getHandler(endpoint)

			w := httptest.NewRecorder()

			// Add two
			obj1 := &Page{
				Content: "foo",
				//				IntValue:10,
			}

			obj2 := &Page{
				Content: "bar",
			}

			collection.Save(obj1)
			collection.Save(obj2)

			req, _ := http.NewRequest("GET", strings.Join([]string{"/api/pages/", obj1.Id.Hex()}, ""), nil)
			//			router.ServeHTTP(w, req)
			handler.ServeHTTP(w,req)

			response := map[string]interface{}{}

			So(w.Code, ShouldEqual, 200)
			err := json.Unmarshal(w.Body.Bytes(), &response)

			So(err, ShouldEqual, nil)

			So(response["Content"], ShouldEqual, "foo")
		})
		Reset(func() {
			conn.Session.DB("bongoz").DropDatabase()

		})
	})
}