FROM docker.friedmicro-lab.org/base

WORKDIR /app

COPY go.mod ./
COPY main.go ./

# Standard library only — no external deps, so no go.sum is needed.
RUN go mod download && \
    go build -o omlx-status .

EXPOSE 8080
CMD ["./omlx-status"]