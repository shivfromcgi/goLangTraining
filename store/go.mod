module cgi.com/goLangTraining/store

go 1.22

require (
	cgi.com/goLangTraining/proto/message_service v0.0.0
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

replace cgi.com/goLangTraining/proto/message_service => ../proto/message_service

replace cgi.com/goLangTraining/src/pkg/storage => ../src/pkg/storage

require (
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
)
