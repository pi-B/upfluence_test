
FROM golang:alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download

RUN go build . 

FROM scratch

WORKDIR /app

COPY --from=builder /app/analysis-api .
EXPOSE 8080

ENTRYPOINT [ "./analysis-api" ]
