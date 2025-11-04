module cgi.com/goLangTraining/src/apps/message-api

go 1.22

replace cgi.com/goLangTraining/src/pkg/storage => ../../pkg/storage

replace cgi.com/goLangTraining/src/pkg/types => ../../pkg/types

replace cgi.com/goLangTraining/src/pkg/middleware => ../../pkg/middleware

require (
	cgi.com/goLangTraining/src/pkg/middleware v0.0.0-00010101000000-000000000000
	cgi.com/goLangTraining/src/pkg/storage v0.0.0-00010101000000-000000000000
	cgi.com/goLangTraining/src/pkg/types v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.0
)

require github.com/google/uuid v1.6.0 // indirect
