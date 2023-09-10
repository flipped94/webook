.PHONY: mock
mock:
	@mockgen -source=pkg/ratelimit/types.go -package=limitmocks -destination=pkg/ratelimit/mocks/ratelimit.mock.go
	@mockgen -source=internal/service/sms/types.go -package=smsmocks -destination=internal/service/sms/mocks/sms.mock.go
	@go mod tidy