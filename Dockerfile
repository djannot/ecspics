FROM google/golang
COPY . /go/src/ecspics/
WORKDIR /go/src/ecspics
RUN go get "github.com/cloudfoundry-community/go-cfenv"
RUN go get "github.com/codegangsta/negroni"
RUN go get "github.com/gorilla/mux"
RUN go get "github.com/gorilla/sessions"
RUN go get "github.com/unrolled/render"
RUN go build .
EXPOSE 80
ENTRYPOINT ["./ecspics"]
CMD ["-Namespace=ns01","-Hostname=10.10.10.1","-EntryPoint=http://namespaces.ecs-local.local:9020"]
