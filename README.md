currency exchange api

go api to get currency exchange rates from frankfurter api and cache with redis

docker:
docker-compose up --build -d
docker cp setup.sql currency-exchange-mssql-1:/tmp/setup.sql
docker exec currency-exchange-mssql-1 /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "sqlPass!223!!" -C -i /tmp/setup.sql

without docker:
go run .

api info:
GET http://localhost:8080/api/v1/health

postman examples:
GET http://localhost:8080/api/v1/rates
GET http://localhost:8080/api/v1/rates/EUR
POST http://localhost:8080/api/v1/convert
{
    "from": "USD",
    "to": "EUR",
    "amount": 100
}
GET http://localhost:8080/api/v1/history?from=USD&limit=5
