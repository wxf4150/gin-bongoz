
 mongodb ODM for go-json-rest   https://github.com/ant0ine/go-json-rest


 modify from  https://github.com/maxwellhealth/bongoz


***sorry  for my bad english!***


#curd api export:
## readlist api
/api/pages?_query={"content":{"$regex":"OO","$options":"i"}}&_perPage=20&_page=1&_sort=
###pagination
* _perPage and _page .   _page is begin with 1, default value _page is 1
* _limit　_skip . when use these parameter , _perPage and _page  will be ignored
### sort
apiurl?_sort=intValue,-dateValue
the sorted fields is split by comma;  - used with  the field  be sorted in reverse order.
### filter-data
* _query
 demo /api/pages?_query={"content":{"$regex":"OO","$options":"i"}}
 the parameter value is  json .is equal the mongo shell db.pages.find({jsonObj-filter}) ;
 but the _query parameter value can't use  field with type of bson.ObjectId
* field parameter
 demo /api/pages?intValue=123&dateValue=unixTime&vObjectId=hexstr
 this fromat support ObjectId field
 the field in the query that type is date should use the conver to  unitTime(seconds since 1970 utc)
 
## create api
	curl /api/pages -d {intValue:123,dateValue:"2016-12-15T03:23:01.109Z"}
## update api
	curl -XPost /api/pages/valueOf_ObjectID -d {intValue:123,dateValue:"2016-12-15T03:23:01.109Z"}
##delete api
	curl -XDelete /api/pages/valueOf_ObjectID
## readByObjectID
	curl /api/pages/valueOf_ObjectID
	
