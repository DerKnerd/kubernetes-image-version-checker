FROM golang:1.15-alpine
WORKDIR /app
COPY . .

RUN go build -o /kubernetes-deployment-version-checker

CMD ["/kubernetes-deployment-version-checker"]