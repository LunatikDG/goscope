# ---- builder: compile the WASM frontend and the server binary -------------
FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd/webdemo && \
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js

RUN CGO_ENABLED=0 go build -o /out/goscope-server ./cmd/serve

# ---- runtime ----------------------------------------------------------------
# "Watch live" shells out to `go run ./examples/<name>` at request time (see
# internal/server/run.go), so the runtime image still needs a Go toolchain and
# the example sources — it isn't a minimal distroless/scratch image. Trimming
# that would mean pre-compiling example binaries at build time instead.
FROM golang:1.26-alpine
WORKDIR /app

RUN adduser -D -u 10001 goscope
# go run's build cache needs a writable $HOME; point it at /tmp for the non-root user.
ENV GOCACHE=/tmp/gocache GOPATH=/tmp/gopath

COPY --from=builder /src/go.mod /src/go.sum ./
COPY --from=builder /src/internal ./internal
COPY --from=builder /src/examples ./examples
COPY --from=builder /src/web ./web
COPY --from=builder /out/goscope-server ./goscope-server

ENV GOSCOPE_ADDR=:8080
EXPOSE 8080
USER goscope

ENTRYPOINT ["./goscope-server"]
