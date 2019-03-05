package bongoz

import (
	"errors"
	"fmt"
	"github.com/go-bongo/bongo"
	"github.com/globalsign/mgo/bson"
	"net/http"
	"strconv"
	"strings"
	"encoding/json"
	"github.com/gin-gonic/gin"
)

type ApiErr struct{
	Error string
}
type ApiOk struct{
	Msg string `example:"ok"`
}

func ErrJson(c *gin.Context,msg string){
	if strings.HasPrefix( msg,"401"){
		c.JSON(401, ApiErr{msg})
		return
	}
	c.JSON(500, ApiErr{msg})
}
func OkJson(c *gin.Context,err error){
	if err!=nil{
		ErrJson(c,err.Error())
	}else{
		c.JSON(200, ApiOk{"ok"})
	}
}
func MsgJson(c *gin.Context,msg string){
	c.JSON(200, ApiOk{msg})
}


type SortConfig struct {
	Field     string
	Direction int
}

type PaginationConfig struct {
	PerPage int
	Sort    []SortConfig
}

type HTTPListResponse struct {
	Pagination *bongo.PaginationInfo
	Data       []interface{}
}

type HTTPSingleResponse struct {
	Data interface{}
}


type ModelFactory func() bongo.Document

//type Middleware struct {
//	ReadOne  alice.Chain
//	ReadList alice.Chain
//	Create   alice.Chain
//	Update   alice.Chain
//	Delete   alice.Chain
//}

type Endpoint struct {
	CollectionName string
	Connection     *bongo.Connection
	Uri            string
	QueryParams    []string
	Pagination     *PaginationConfig
	Factory        ModelFactory
//	Middleware     *Middleware

	AllowFullQuery bool
	DisableWrites  bool
}

func NewEndpoint(uri string, connection *bongo.Connection, collectionName string) *Endpoint {
	endpoint := new(Endpoint)
	endpoint.Uri = uri
	endpoint.Connection = connection
	endpoint.CollectionName = collectionName
	endpoint.Pagination = &PaginationConfig{}
//	endpoint.Middleware = new(Middleware)
	return endpoint
}

func (e *Endpoint) Register(r *gin.RouterGroup) {
	r.GET(e.Uri,e.HandleReadList)
	r.GET(e.Uri+"/:id",e.HandleReadOne)
	if !e.DisableWrites {
		r.POST(e.Uri,e.HandleCreate)
		r.POST(e.Uri+"/:id",e.HandleUpdate)
		r.DELETE(e.Uri+"/:id", e.HandleDelete)
	}
}




// Register the endpoint to the http root handler. Use GetRouter() for more flexibility
//func (e *Endpoint) Register(r *mux.Router) {
//	e.registerRoutes(r)
//}

// Get the mux router that can be plugged in as an http handler.
// Gives more flexibility than just using the Register() method which
// registers the router directly on the http root handler.
// Use this is you want to use a subroute, a custom http.Server instance, etc
//func (e *Endpoint) GetRouter() *mux.Router {
//	r := mux.NewRouter()
//	e.registerRoutes(r)
//	return rparsed
//}
//
//func (e *Endpoint) registerRoutes(r *mux.Router) {
//	r.Handle(e.Uri, e.Middleware.ReadList.ThenFunc(e.HandleReadList)).Methods("GET")
//	r.Handle(e.Uri+"/{id}", e.Middleware.ReadOne.ThenFunc(e.HandleReadOne)).Methods("GET")
//
//	if !e.DisableWrites {
//		r.Handle(e.Uri, e.Middleware.Create.ThenFunc(e.HandleCreate)).Methods("POST")
//
//		r.Handle(e.Uri+"/{id}", e.Middleware.Update.ThenFunc()).Methods("PUT")
//		r.Handle(e.Uri+"/{id}", e.Middleware.Delete.ThenFunc(e.HandleDelete)).Methods("DELETE")
//	}
//
//}

//func methodsFromMethod(method string) []string {
//	if method == "*" || method == "all" {
//		return []string{"ReadOne", "ReadList", "Create", "Update", "Delete"}
//	} else if method == "write" {
//		return []string{"Create", "Update", "Delete"}
//	} else if method == "read" {
//		return []string{"ReadOne", "ReadList"}
//	} else {
//		return []string{method}
//	}
//}
//func (e *Endpoint) SetMiddleware(method string, chain alice.Chain) *Endpoint {
//	methods := methodsFromMethod(method)
//	for _, m := range methods {
//		switch m {
//		case "ReadOne":
//			e.Middleware.ReadOne = chain
//		case "ReadList":
//			e.Middleware.ReadList = chain
//		case "Create":
//			e.Middleware.Create = chain
//		case "Update":
//			e.Middleware.Update = chain
//		case "Delete":
//			e.Middleware.Delete = chain
//
//		}
//	}
//	return e
//}

func handleError(c *gin.Context) {
	var err error
	if r := recover(); r != nil {
		// panic(r)
		// return
		if e, ok := r.(error); ok {
			if e.Error() == "EOF" {
				err = errors.New("Lost database connection unexpectedly")
			} else {
				err = e
			}

		} else if e, ok := r.(string); ok {
			err = errors.New(e)
		} else {
			err = errors.New(fmt.Sprint(r))
		}
		OkJson(c,err)
	}
}

// Handle a "ReadList" request, including parsing pagination, query string, etc
func (e *Endpoint) HandleReadList(c *gin.Context) {
	w:=c.Writer.(http.ResponseWriter)
	req:=c.Request
	defer handleError(c)
	w.Header().Set("Content-Type", "application/json")
	var err error
	//var code int

	// Get the query
	query, err := e.getQuery(req)

	if err != nil {
		OkJson(c,err)
		return
	}

	connection := e.Connection

	results := connection.Collection(e.CollectionName).Find(query)

	defer results.Free()

	// Default pagination is 50
	if e.Pagination.PerPage == 0 {
		e.Pagination.PerPage = 50
	}

	perPage := e.Pagination.PerPage
	limit := 0
	skip := 0
	page := 1

	// Allow override with query vars
	perPageParam := req.URL.Query().Get("_perPage")
	pageParam := req.URL.Query().Get("_page")

	// Allow support for limit and skip with no pagination
	limitParam := req.URL.Query().Get("_limit")
	skipParam := req.URL.Query().Get("_skip")

	paginate := true

	if len(limitParam) > 0 {
		paginate = false
		converted, err := strconv.Atoi(limitParam)

		if err == nil && converted > 0 {
			limit = converted
		}
	}

	if len(skipParam) > 0 {

		converted, err := strconv.Atoi(skipParam)
		paginate = false
		if err == nil && converted >= 0 {
			skip = converted
		}
	}

	var pageInfo *bongo.PaginationInfo

	if limit > 0 {
		results.Query.Limit(limit).Skip(skip)
		pageInfo = &bongo.PaginationInfo{}
		pageInfo.TotalPages = 1
		pageInfo.Current = 1
	}

	if len(perPageParam) > 0 {
		converted, err := strconv.Atoi(perPageParam)
		// Hard limit to 500 so people can break it
		if err == nil && converted > 0 && converted < 500 {
			perPage = converted
		}
	}

	if len(pageParam) > 0 {
		converted, err := strconv.Atoi(pageParam)

		if err == nil && converted >= 1 {
			page = converted
		}
	}

	var total int
	if paginate {
		pageInfo, err = results.Paginate(perPage, page)
		if err != nil {
			panic(err)
		}

		total = pageInfo.RecordsOnPage

	} else {
		total, err = results.Query.Count()
		if err != nil {
			panic(err)
		}

		pageInfo.RecordsOnPage = total
		pageInfo.TotalRecords = total
		pageInfo.PerPage = total
		if total == 0 {
			pageInfo.TotalPages = 0
		}
	}

	sortParam := req.URL.Query().Get("_sort")
	if len(sortParam) > 0 {
		sortFields := strings.Split(sortParam, ",")
		results.Query.Sort(sortFields...)
	}

	response := make([]interface{}, total)
	for i := 0; i < total; i++ {
		res := e.Factory()
		results.Next(res)
		response[i] = res

	}

	httpResponse := &HTTPListResponse{pageInfo, response}
	c.JSON(200,httpResponse)

	//encoder := ojson.NewEncoder(w)
	//err = encoder.Encode(httpResponse)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	io.WriteString(w, NewErrorResponse(err).ToJSON())
	//}
}

func (e *Endpoint) HandleReadOne(c *gin.Context) {

	w:=c.Writer.(http.ResponseWriter)
	//req:=c.Request
	defer handleError(c)
	var err error

	id:=c.Param("id")
	if len(id) == 0 || !bson.IsObjectIdHex(id) {
		http.Error(w, "Invalid object ID", http.StatusBadRequest)
		return
	}
	instance := e.Factory()
	err = e.Connection.Collection(e.CollectionName).FindById(bson.ObjectIdHex(id), instance)

	if err != nil {
		OkJson(c,err)
		return
	}
	c.JSON(200,instance)

	//httpResponse := &HTTPSingleResponse{instance}
	//encoder := ojson.NewEncoder(w)
	//err = encoder.Encode(httpResponse)
	//
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	io.WriteString(w, NewErrorResponse(err).ToJSON())
	//}
}

func (e *Endpoint) HandleCreate(c *gin.Context) {
	req := c.Request
	defer handleError(c)
	var err error

	decoder := json.NewDecoder(req.Body)

	obj := e.Factory()
	// Instantiate diff tracker
	if trackable, ok := obj.(bongo.Trackable); ok {
		trackable.GetDiffTracker().Reset()
	}

	err = decoder.Decode(obj)
	if err == nil {
		err = e.Connection.Collection(e.CollectionName).Save(obj)
		if err == nil {
			err = e.Connection.Collection(e.CollectionName).FindById(obj.GetId(), obj)
			if err==nil{
				c.JSON(201, obj)
			}
		}
	}
	if err != nil {
		OkJson(c, err)
	}
}

func (e *Endpoint) HandleUpdate(c *gin.Context) {
	//w:=c.Writer.(http.ResponseWriter)
	//req:=c.Request
	defer handleError(c)

	var err error

	id:=c.Param("id")

	if len(id) == 0 || !bson.IsObjectIdHex(id) {
		OkJson(c,errors.New("Invalid Object ID:"+id))
		return
	}

	// Execute the find
	instance := e.Factory()

	err = e.Connection.Collection(e.CollectionName).FindById(bson.ObjectIdHex(id), instance)
	if err != nil {
		OkJson(c,err)
		return
	}

	if trackable, ok := instance.(bongo.Trackable); ok {
		trackable.GetDiffTracker().Reset()
	}

	// Save the ID and reapply it afterward, so we do not allow the http request to modify the ID
	actualId := instance.GetId()

	err=c.BindJSON(instance)
	if err != nil {
		OkJson(c,err)
		return
	}
	instance.SetId(actualId)
	//if tt, ok := instance.( bongo.TimeModifiedTracker); ok {
	//	tt.SetModified(time.Now())
	//}
	err = e.Connection.Collection(e.CollectionName).Save(instance)
	if err != nil {
		OkJson(c,err)
		return
	}
	c.JSON(200,instance)

	//httpResponse := &HTTPSingleResponse{instance}
	//encoder := ojson.NewEncoder(w)
	//err = encoder.Encode(httpResponse)
	//
	//if err != nil {
	//	panic(err)
	//}

}

func (e *Endpoint) HandleDelete(c *gin.Context) {
	//req:=c.Request
	defer handleError(c)
	var err error

	id:=c.Param("id")

	if len(id) == 0 || !bson.IsObjectIdHex(id) {
		OkJson(c,errors.New("Invalid Object ID"))
		return
	}
	// Execute the find
	instance := e.Factory()

	// Use a FindOne instead of FindById since the query filters may need
	// to add additional parameters to the search query, aside from just ID.
	// Error here is just if there is no document
	collection := e.Connection.Collection(e.CollectionName)

	err = collection.FindById(bson.ObjectIdHex(id), instance)
	if err != nil {
		OkJson(c,err)
		return
	}
	err = collection.DeleteDocument(instance)
	OkJson(c,err)

}
