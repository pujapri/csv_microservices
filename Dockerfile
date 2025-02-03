FROM golang:1.23.3

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./main.go

CMD ["C:\Users\puja.priyanshu\Desktop\Practise\.vscode\Go_Practise\csv-microservice\main.go"]
