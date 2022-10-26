export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
.DEFAULT_GOAL := deploy

deploy:
	go build -o function
	zip -r function.zip function
	aws lambda update-function-code --function-name "go-lambda-test" --zip-file fileb://function.zip --region="eu-west-1" | jq .  