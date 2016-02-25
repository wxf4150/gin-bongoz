package bongoz

import (
//	"github.com/DailyFeats/dpl/models/traits"

	"github.com/maxwellhealth/bongo"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
	// "net/url"
	"errors"
	"reflect"
	"github.com/ant0ine/go-json-rest/rest"
)

func getConnection() *bongo.Connection {
	conf := &bongo.Config{
		ConnectionString: "192.168.9.100",
		Database:         "bongoz",
	}

	conn, err := bongo.Connect(conf)

	if err != nil {
		panic(err)
	}

	return conn
}
func getHandler(endpoint *Endpoint)(http.Handler){
	api := rest.NewApi()
	router,err1 :=rest.MakeRouter(endpoint.GetJRouters()...)
	if err1 != nil {
		panic(err1)
	}
//	So(err1, ShouldEqual, nil)
	api.SetApp(router)
	handler:=api.MakeHandler()
	return handler
}

func errorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		// h.ServeHTTP(w, r)
	})
}

type singleResponse struct {
	Data map[string]interface{}
}

type Page struct {
	bongo.DocumentBase `bson:",inline"`
	Content            string
	IntValue           int                    `bson:"intValue"`
	DateValue          time.Time              `bson:"dateValue"`
	ArrValue           []string               `bson:"arrValue"`
	IdArr              []bson.ObjectId        `bson:"idArr"`
	IdValue            bson.ObjectId          `json:",omitempty" bson:"idValue,omitempty"`
	RandomMap          map[string]interface{} `bson:"randomMap"`
}

func Factory() bongo.Document {
	return &Page{}
}

type HistoricalPage struct {
	Page              `bson:",inline"`
//	traits.Historical `bson:",inline"`
	OtherVal          string
	diffTracker       *bongo.DiffTracker
}

func (f *HistoricalPage) GetDiffTracker() *bongo.DiffTracker {
	v := reflect.ValueOf(f.diffTracker)
	if !v.IsValid() || v.IsNil() {
		f.diffTracker = bongo.NewDiffTracker(f)
	}

	return f.diffTracker
}

func HistoricalFactory() bongo.Document {
	return &HistoricalPage{}
}

type validatedModel struct {
	bongo.DocumentBase `bson:",inline"`
	Content            string `json:"content"`
}

func ValidFactory() bongo.Document {
	return &validatedModel{}
}

func (v *validatedModel) Validate(collection *bongo.Collection) []error {
	ret := []error{}
	if !bongo.ValidateRequired(v.Content) {
		ret = append(ret, errors.New("Content is required"))
	}

	return ret
}
