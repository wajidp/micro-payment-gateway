

.PHONY: swagger
swagger: ## generate swagger file
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate -i /local/docs/swagger.yaml -g html -o /local/swagger-docs
