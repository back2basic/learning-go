# rss agg

## install godotenv

- go get github.com/joho/godotenv
- go mod vendor
- go mod tidy
- go mod vendor

## postgress

- docker run --name postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres
