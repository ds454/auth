FROM    golang:1.27.3-alpine3.24 AS build
WORKDIR /build
COPY    .   .

# Install bash and protoc
RUN     apk update && \
        apk add \
            protoc \
            protobuf-dev

# Install golang plugins
RUN     chmod +x ./.devcontainer/install-proto-plugins.sh && \
        ./.devcontainer/install-proto-plugins.sh

# Transpile protobuf to go
RUN     chmod +x ./local-dev/apply-proto-go.sh && \
        ./local-dev/apply-proto-go.sh

# Compile project
RUN     go build -v -x -o app

# Actual image
FROM    alpine:3.24
COPY    --from=build /build/app /app
ENTRYPOINT [ "/app" ]