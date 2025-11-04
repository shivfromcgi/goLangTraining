module cgi.com/goLangTraining/src/apps/message-api

go 1.22

replace cgi.com/goLangTraining/src/pkg/storage => ../../pkg/storage

replace cgi.com/goLangTraining/src/pkg/types => ../../pkg/types

require (
	cgi.com/goLangTraining/src/pkg/storage v0.0.0-00010101000000-000000000000
	cgi.com/goLangTraining/src/pkg/types v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
)
