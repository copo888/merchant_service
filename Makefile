t := boadmin
lang:
	easyi18n generate --pkg=locales ./locales ./locales/locales.go

api:
	goctl api go -api $(t)/$(t).api -dir $(t) -style gozero --home template/1.2.4-cli

run:
	go run $(t)/$(t).go -f $(t)/etc/$(t)-api.yaml -env $(t)/etc/.env

mer:
	go run merchant/merchant.go -f merchant/etc/merchant-api.yaml -env merchant/etc/.env

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/boadmin_service boadmin/boadmin.go