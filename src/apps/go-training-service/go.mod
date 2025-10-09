module cgi.com/goLangTraining/src/apps/go-training-service

go 1.21

require (
	cgi.com/goLangTraining/src/pkg/storage v0.0.0
	github.com/google/uuid v1.6.0
)

replace cgi.com/goLangTraining/src/pkg/storage => ../../pkg/storage
