package main

import (
  "bytes"
  "crypto/tls"
  "encoding/json"
  "encoding/xml"
  "flag"
  "log"
  "net/http"
  "net/url"
  "os"
  "strconv"
  "strings"
  "time"
  cfenv "github.com/cloudfoundry-community/go-cfenv"
  "github.com/codegangsta/negroni"
  "github.com/gorilla/mux"
  "github.com/gorilla/sessions"
  "github.com/unrolled/render"
)

var rendering *render.Render
var store = sessions.NewCookieStore([]byte("session-key"))

func contains(dict map[string]string, i string) bool {
  if _, ok := dict[i]; ok {
    return true
  } else {
    return false
  }
}

func int64toString(value int64) (string) {
	return strconv.FormatInt(value, 10)
}

func int64InSlice(i int64, list []int64) bool {
  for _, value := range list {
        if value == i {
            return true
        }
    }
    return false
}

type appError struct {
	err error
	status int
	json string
	template string
	binding interface{}
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  if e := fn(w, r); e != nil {
		log.Print(e.err)
		if e.status != 0 {
			if e.json != "" {
				rendering.JSON(w, e.status, e.json)
			} else {
				rendering.HTML(w, e.status, e.template, e.binding)
			}
		}
  }
}

func RecoverHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if err := recover(); err != nil {
        log.Printf("panic: %+v", err)
        http.Error(w, http.StatusText(500), 500)
      }
    }()

    next.ServeHTTP(w, r)
  }
	return http.HandlerFunc(fn)
}

func LoginMiddleware(h http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/login" || strings.HasPrefix(r.URL.Path, "/app") {
      h.ServeHTTP(w, r)
    } else {
      session, err := store.Get(r, "session-name")
      if err != nil {
        rendering.HTML(w, http.StatusInternalServerError, "error", http.StatusInternalServerError)
      }
      if _, ok := session.Values["AccessKey"]; ok {
        h.ServeHTTP(w, r)
      } else {
        http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
        //rendering.HTML(w, http.StatusOK, "login", nil)
      }
    }
  })
}

type UploadData struct {
  Hostname string
  DockerHost string
  Buckets []string
}

type ECS struct {
  Hostname string `json:"hostname"`
  EndPoint string `json:"endpoint"`
  Namespace string `json:"namespace"`
}

var hostname string
var ecs ECS

func main() {
  var port = ""
  _, err := cfenv.Current()
  if(err != nil) {
    port = "80"
    endPointPtr := flag.String("EndPoint", "", "The Amazon S3 endpoint")
    namespacePtr := flag.String("Namespace", "", "The ViPR namespace if used in the Object Base URL")
    hostnamePtr := flag.String("Hostname", "", "The ECS hostname or IP address")
    flag.Parse()
    ecs = ECS{
      Hostname: *hostnamePtr,
      EndPoint: *endPointPtr,
      Namespace: *namespacePtr,
    }
  } else {
    port = os.Getenv("PORT")
    ecs = ECS{
      Hostname: os.Getenv("HOSTNAME"),
      EndPoint: os.Getenv("ENDPOINT"),
      Namespace: os.Getenv("NAMESPACE"),
    }
  }

  hostname, _ = os.Hostname()

  // See http://godoc.org/github.com/unrolled/render
  rendering = render.New(render.Options{Directory: "app/templates"})

  // See http://www.gorillatoolkit.org/pkg/mux
  router := mux.NewRouter()
  router.HandleFunc("/", Index)
  router.Handle("/api/v1/buckets", appHandler(Buckets)).Methods("GET")
  router.HandleFunc("/api/v1/hostname", Hostname).Methods("GET")
  router.HandleFunc("/api/v1/ecs", Ecs).Methods("GET")
  router.Handle("/api/v1/createbucket", appHandler(CreateBucket)).Methods("POST")
	router.Handle("/api/v1/uploadpicture", appHandler(UploadPicture)).Methods("POST")
  router.Handle("/api/v1/search", appHandler(Search)).Methods("POST")
  router.HandleFunc("/login", Login)
  router.HandleFunc("/logout", Logout)
  router.PathPrefix("/app/").Handler(http.StripPrefix("/app/", http.FileServer(http.Dir("app"))))

	n := negroni.Classic()
	n.UseHandler(RecoverHandler(LoginMiddleware(router)))
	n.Run(":" + port)

	log.Printf("Listening on port " + port)
}

type UserSecretKeysResult struct {
  XMLName xml.Name `xml:"user_secret_keys"`
  SecretKey1 string `xml:"secret_key_1"`
  SecretKey2 string `xml:"secret_key_2"`
}

type UserSecretKeyResult struct {
  XMLName xml.Name `xml:"user_secret_key"`
  SecretKey string `xml:"secret_key"`
}

func Login(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    r.ParseForm()
    user := r.FormValue("user")
    password := r.FormValue("password")
    tr := &http.Transport{
      TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}
    req, _ := http.NewRequest("GET", "https://" + ecs.Hostname + ":4443/login", nil)
    req.SetBasicAuth(user, password)
    resp, err := client.Do(req)
    log.Print(req)
    log.Print(resp)
    if err != nil{
        log.Print(err)
    }
    if resp.StatusCode == 401 {
      rendering.HTML(w, http.StatusOK, "login", "Check your crententials and that you're allowed to generate a secret key on ECS")
    } else {
      req, _ = http.NewRequest("GET", "https://" + ecs.Hostname + ":4443/object/secret-keys", nil)
      headers := map[string][]string{}
      headers["X-Sds-Auth-Token"] = []string{resp.Header.Get("X-Sds-Auth-Token")}
      req.Header = headers
      resp, err = client.Do(req)
      if err != nil{
          log.Print(err)
      }
      buf := new(bytes.Buffer)
      buf.ReadFrom(resp.Body)
      secretKey := ""
      userSecretKeysResult := &UserSecretKeysResult{}
      xml.NewDecoder(buf).Decode(userSecretKeysResult)
      secretKey = userSecretKeysResult.SecretKey1
      if secretKey == "" {
        req, _ = http.NewRequest("POST", "https://" + ecs.Hostname + ":4443/object/secret-keys", bytes.NewBufferString("<secret_key_create_param></secret_key_create_param>"))
        headers["Content-Type"] = []string{"application/xml"}
        req.Header = headers
        resp, err = client.Do(req)
        if err != nil{
            log.Print(err)
        }
        buf = new(bytes.Buffer)
        buf.ReadFrom(resp.Body)
        userSecretKeyResult := &UserSecretKeyResult{}
        xml.NewDecoder(buf).Decode(userSecretKeyResult)
        secretKey = userSecretKeyResult.SecretKey
      }
      session, err := store.Get(r, "session-name")
      if err != nil {
        rendering.HTML(w, http.StatusInternalServerError, "error", http.StatusInternalServerError)
      }
      session.Values["AccessKey"] = user
      session.Values["SecretKey"] = secretKey
      err = sessions.Save(r, w)
      if err != nil {
        rendering.HTML(w, http.StatusInternalServerError, "error", http.StatusInternalServerError)
      }
      http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
    }
  } else {
    rendering.HTML(w, http.StatusOK, "login", nil)
  }
}

func Logout(w http.ResponseWriter, r *http.Request) {
  session, err := store.Get(r, "session-name")
  if err != nil {
    rendering.HTML(w, http.StatusInternalServerError, "error", http.StatusInternalServerError)
  }
  delete(session.Values, "AccessKey")
  delete(session.Values, "SecretKey")
  err = sessions.Save(r, w)

  http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func Index(w http.ResponseWriter, r *http.Request) {
  rendering.HTML(w, http.StatusOK, "index", nil)
}

func Hostname(w http.ResponseWriter, r *http.Request) {
  rendering.JSON(w, http.StatusOK, hostname)
}

func Ecs(w http.ResponseWriter, r *http.Request) {
  rendering.JSON(w, http.StatusOK, ecs)
}

func Buckets(w http.ResponseWriter, r *http.Request) *appError {
  session, err := store.Get(r, "session-name")
  if err != nil {
    return &appError{err: err, status: http.StatusInternalServerError, json: http.StatusText(http.StatusInternalServerError)}
  }
  s3 := S3{
    EndPointString: ecs.EndPoint,
    AccessKey: session.Values["AccessKey"].(string),
    SecretKey: session.Values["SecretKey"].(string),
    Namespace: ecs.Namespace,
  }
  response, _ := s3Request(s3, "", "GET", "/", make(map[string][]string), "")
  listBucketsResp := &ListBucketsResp{}
  xml.NewDecoder(strings.NewReader(response.Body)).Decode(listBucketsResp)
  buckets := []string{}
  for _, bucket := range listBucketsResp.Buckets {
    buckets = append(buckets, bucket.Name)
  }
  rendering.JSON(w, http.StatusOK, buckets)

  return nil
}

func CreateBucket(w http.ResponseWriter, r *http.Request) *appError {
  session, err := store.Get(r, "session-name")
  if err != nil {
    return &appError{err: err, status: http.StatusInternalServerError, json: http.StatusText(http.StatusInternalServerError)}
  }
  s3 := S3{
    EndPointString: ecs.EndPoint,
    AccessKey: session.Values["AccessKey"].(string),
    SecretKey: session.Values["SecretKey"].(string),
    Namespace: ecs.Namespace,
  }

  decoder := json.NewDecoder(r.Body)
  var s map[string]string
  err = decoder.Decode(&s)
  if err != nil {
    return &appError{err: err, status: http.StatusBadRequest, json: "Can't decode JSON data"}
  }
  bucketName := s["bucket"]

  createBucketHeaders := map[string][]string{}
  createBucketHeaders["Content-Type"] = []string{"application/xml"}
  createBucketHeaders["x-emc-is-stale-allowed"] = []string{"true"}
  createBucketHeaders["x-emc-metadata-search"] = []string{"ObjectName,x-amz-meta-image-width;Integer,x-amz-meta-image-height;Integer,x-amz-meta-gps-latitude;Decimal,x-amz-meta-gps-longitude;Decimal"}

  createBucketResponse, _ := s3Request(s3, bucketName, "PUT", "/", createBucketHeaders, "")
  if createBucketResponse.Code == 200 {
    enableBucketCorsHeaders := map[string][]string{}
    enableBucketCorsHeaders["Content-Type"] = []string{"application/xml"}
    corsConfiguration := `
      <CORSConfiguration>
       <CORSRule>
         <AllowedOrigin>*</AllowedOrigin>
         <AllowedHeader>*</AllowedHeader>
         <ExposeHeader>x-amz-meta-image-width</ExposeHeader>
         <ExposeHeader>x-amz-meta-image-height</ExposeHeader>
         <ExposeHeader>x-amz-meta-gps-latitude</ExposeHeader>
         <ExposeHeader>x-amz-meta-gps-longitude</ExposeHeader>
         <AllowedMethod>HEAD</AllowedMethod>
         <AllowedMethod>GET</AllowedMethod>
         <AllowedMethod>PUT</AllowedMethod>
         <AllowedMethod>POST</AllowedMethod>
         <AllowedMethod>DELETE</AllowedMethod>
       </CORSRule>
      </CORSConfiguration>
    `
    enableBucketCorsResponse, _ := s3Request(s3, bucketName, "PUT", "/?cors", enableBucketCorsHeaders, corsConfiguration)
    if enableBucketCorsResponse.Code == 200 {
      rendering.JSON(w, http.StatusOK, struct {
        CorsConfiguration string `json:"cors_configuration"`
        Bucket string `json:"bucket"`
      } {
        CorsConfiguration: corsConfiguration,
        Bucket: bucketName,
      })
    } else {
      return &appError{err: err, status: http.StatusBadRequest, json: "Bucket created, but CORS can't be enabled"}
    }
  } else {
    return &appError{err: err, status: http.StatusBadRequest, json: "Bucket can't be created"}
  }
  return nil
}

func UploadPicture(w http.ResponseWriter, r *http.Request) *appError {
  session, err := store.Get(r, "session-name")
  if err != nil {
    return &appError{err: err, status: http.StatusInternalServerError, json: http.StatusText(http.StatusInternalServerError)}
  }
  s3 := S3{
    EndPointString: ecs.EndPoint,
    AccessKey: session.Values["AccessKey"].(string),
    SecretKey: session.Values["SecretKey"].(string),
    Namespace: ecs.Namespace,
  }

  decoder := json.NewDecoder(r.Body)
  var s map[string]string
  err = decoder.Decode(&s)
  if err != nil {
    return &appError{err: err, status: http.StatusBadRequest, json: "Can't decode JSON data"}
  }
  bucketName := s["bucket"]
  retention := s["retention"]
  fileName := s["file_name"]
  fileSize := s["file_size"]
  imageWidth := s["image_width"]
  imageHeight := s["image_height"]
  gpsLatitude := s["gps_latitude"]
  gpsLongitude := s["gps_longitude"]
  datetime := s["datetime"]

  contentType := "binary/octet-stream"
  headers := make(map[string][]string)
  headers["Content-Length"] = []string{fileSize}
  headers["Content-Type"] = []string{contentType}
  if retention != "" {
    i, err := strconv.Atoi(retention)
    if err != nil {
      return &appError{err: err, status: http.StatusBadRequest, json: "Can't use this retention value"}
    }
    headers["x-emc-retention-period"] = []string{strconv.Itoa(i * 24 * 3600)}
  }
  headers["x-amz-meta-image-width"] = []string{imageWidth}
  headers["x-amz-meta-image-height"] = []string{imageHeight}
  if gpsLatitude != "" {
    headers["x-amz-meta-gps-latitude"] = []string{gpsLatitude}
  }
  if gpsLongitude != "" {
    headers["x-amz-meta-gps-longitude"] = []string{gpsLongitude}
  }
  if datetime != "" {
    headers["x-amz-meta-datetime"] = []string{datetime}
  }
  preparedS3Request, _ := prepareS3Request(s3, bucketName, "PUT", "/pictures/" + fileName, headers, true)
  headersToSend := ""
  for k, v := range headers {
    headersToSend += `request.setRequestHeader('` + k + `','` + v[0] +`');
                      `
  }

  rendering.JSON(w, http.StatusOK, struct {
    Headers map[string][]string `json:"headers"`
    Url string `json:"url"`
  } {
    Headers: headers,
    Url: preparedS3Request.Url,
  })
  return nil
}

type Query struct {
  Bucket string `json:"search_bucket"`
  Width string `json:"search_width"`
  Height string `json:"search_height"`
  Area bool `json:"search_area"`
  SWLatitude string `json:"search_sw_latitude"`
  SWLongitude string `json:"search_sw_longitude"`
  NELatitude string `json:"search_ne_latitude"`
  NELongitude string `json:"search_ne_longitude"`
}

type Picture struct {
  MediaUrl string
  Key string
  Metadatas map[string]string
  DeleteRequestHeaders map[string][]string
  DeleteRequestUrl string
}

func Search(w http.ResponseWriter, r *http.Request) *appError {
  session, err := store.Get(r, "session-name")
  if err != nil {
    return &appError{err: err, status: http.StatusInternalServerError, json: http.StatusText(http.StatusInternalServerError)}
  }
  s3 := S3{
    EndPointString: ecs.EndPoint,
    AccessKey: session.Values["AccessKey"].(string),
    SecretKey: session.Values["SecretKey"].(string),
    Namespace: ecs.Namespace,
  }
  decoder := json.NewDecoder(r.Body)
  var query Query
  err = decoder.Decode(&query)
  if err != nil {
    return &appError{err: err, status: http.StatusBadRequest, json: "Can't decode JSON data"}
  }

  imageWidth := "0"
  imageHeight := "0"
  if(query.Width != "") {
    imageWidth = query.Width
  }
  if(query.Height != "") {
    imageHeight = query.Height
  }
  path := ""
  if query.Area {
    //swLongitude, _ := strconv.ParseFloat(query.SWLongitude, 64)
    //neLongitude, _ := strconv.ParseFloat(query.NELongitude, 64)
    //if swLongitude < neLongitude {
      path = "/?query=x-amz-meta-image-width%20>%20" + imageWidth + "%20and%20x-amz-meta-image-height%20>%20" + imageHeight + "%20and%20x-amz-meta-gps-latitude%20>%20" + query.SWLatitude + "%20and%20x-amz-meta-gps-latitude%20<%20" + query.NELatitude + "%20and%20x-amz-meta-gps-longitude%20>%20" + query.SWLongitude + "%20and%20x-amz-meta-gps-longitude%20<%20" + query.NELongitude + "&attributes=Retention"
    //} else {
      //path = "/?query=x-amz-meta-image-width%20>%20" + imageWidth + "%20and%20x-amz-meta-image-height%20>%20" + imageHeight + "%20and%20x-amz-meta-gps-latitude%20>%20" + query.SWLatitude + "%20and%20x-amz-meta-gps-latitude%20<%20" + query.NELatitude + "%20and(%20x-amz-meta-gps-longitude%20>%20" + query.SWLongitude + "%20or%20x-amz-meta-gps-longitude%20<%20" + query.NELongitude + ")&attributes=Retention"
    //}
  } else {
    path = "/?query=x-amz-meta-image-width%20>%20" + imageWidth + "%20and%20x-amz-meta-image-height%20>%20" + imageHeight + "&attributes=Retention"
  }
  bucketQueryResponse, err := s3Request(s3, query.Bucket, "GET", path, make(map[string][]string), "")
  if err != nil {
    return &appError{err: err, status: http.StatusInternalServerError, json: http.StatusText(http.StatusInternalServerError)}
  } else {
    bucketQueryResult := &BucketQueryResult{}
    xml.NewDecoder(strings.NewReader(bucketQueryResponse.Body)).Decode(bucketQueryResult)
    var pictures []Picture
    if len(bucketQueryResult.EntryLists) > 0 {
      for _, item := range bucketQueryResult.EntryLists {
        if item.ObjectName[len(item.ObjectName)-1:] != "/" {
          expires := time.Now().Add(time.Second*3600)
          headers := make(map[string][]string)
          preparedS3Request, _ := prepareS3Request(s3, query.Bucket, "GET", item.ObjectName + "?Expires=" + strconv.FormatInt(expires.Unix(), 10), headers, true)
          v := url.Values{}
          v = preparedS3Request.Params
          deleteHeaders := make(map[string][]string)
          preparedS3DeleteRequest, _ := prepareS3Request(s3, query.Bucket, "DELETE", item.ObjectName, deleteHeaders, true)
          metadatas := map[string]string{}
          for _, metadata := range item.Metadatas {
            metadatas[metadata.Key] = metadata.Value
          }
          pictures = append(pictures, Picture{MediaUrl: strings.Split(preparedS3Request.Url, "?")[0] + "?" + v.Encode(), Key: item.ObjectName, DeleteRequestHeaders: deleteHeaders, DeleteRequestUrl: preparedS3DeleteRequest.Url, Metadatas: metadatas})
        }
      }
    } else {
      return &appError{err: err, status: http.StatusBadRequest, json: "The specified search didn't return any result"}
    }
    rendering.JSON(w, http.StatusOK, pictures)
    return nil
  }
}
