module cgi.com/goLangTraining/src/apps/message-grpc

go 1.24.0

replace cgi.com/goLangTraining/proto/message_service => ../../../proto/message_service

replace cgi.com/goLangTraining/src/pkg/storage => ../../pkg/storage

replace cgi.com/goLangTraining/src/pkg/types => ../../pkg/types

require (
	cgi.com/goLangTraining/proto/message_service v0.0.0-00010101000000-000000000000
	cgi.com/goLangTraining/src/pkg/storage v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	cgi.com/goLangTraining/src/pkg/types v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
)
