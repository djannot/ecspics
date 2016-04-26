FROM google/golang
WORKDIR /go/src
RUN git clone https://github.com/djannot/ecspics.git
WORKDIR /go/src/ecspics
RUN go get "github.com/cloudfoundry-community/go-cfenv"
RUN go get "github.com/codegangsta/negroni"
RUN go get "github.com/gorilla/mux"
RUN go get "github.com/gorilla/sessions"
RUN go get "github.com/unrolled/render"
RUN go build .
