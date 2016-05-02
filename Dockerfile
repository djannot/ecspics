FROM google/golang
COPY . /go/src/ecspics/
WORKDIR /go/src/ecspics
RUN go get "github.com/cloudfoundry-community/go-cfenv"
RUN go get "github.com/codegangsta/negroni"
RUN go get "github.com/gorilla/mux"
RUN go get "github.com/gorilla/sessions"
RUN go get "github.com/unrolled/render"
RUN go build .
# uncomment to set environment in dockerfile, otherwise use -e flag
#ENV PORT 80
#ENV HOSTNAME 10.10.10.1
#ENV ENDPOINT http://ecs-vip.local.net:9020
#ENV NAMESPACE ns01
ENTRYPOINT ["./ecspics"]
