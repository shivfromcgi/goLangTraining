module cgi.com/goLangTraining

go 1.22

replace cgi.com/goLangTraining/src/pkg/storage => ./src/pkg/storage

require (
	cgi.com/goLangTraining/src/pkg/storage v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
)
