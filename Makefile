GIT_VERSION = $(shell git describe --always)

# build:
# 	docker build -t civelocity/${REPOSITORY}:${GIT_VERSION} .
# 	docker tag civelocity/${REPOSITORY}:${GIT_VERSION} civelocity/${REPOSITORY}:latest
install:
	@docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
	golang:1.10 \
	scripts/install-deps.sh

build:
	@docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/parameters-aws \
	--workdir /go/src/github.com/velocity-ci/parameters-aws \
	golang:1.10-alpine \
	scripts/build.sh

# publish: 
# 	docker run --rm \
# 	--volume ${CURDIR}:/app \
# 	--workdir /app \
# 	python:3 \
# 	scripts/publish.py velocity-ci parameters-aws-credentials dist/aws-credentials
	# docker push civelocity/${REPOSITORY}:${GIT_VERSION}
	# docker push civelocity/${REPOSITORY}:latest

