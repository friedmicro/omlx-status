FROM docker.friedmicro-lab.org/base AS build

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go mod download
RUN go build -o /opt/omlx-status

FROM docker.friedmicro-lab.org/base-slim
COPY --from=build /opt/omlx-status /opt/omlx-status

EXPOSE 8080
CMD ["/opt/omlx-status"]
