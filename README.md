# authservice
An user authentication and authorization service.

- To use, set the following env variables in a `.env` file
```
PORT=9000
PG_PORT=7000
PG_HOST=localhost
PG_USERNAME=admin
PG_PASSWORD=admin
PG_DATABASE=authservice
REDIS_PORT=7002
REDIS_HOST=localhost
REDIS_USERNAME=default
TOKEN_SECRET=testToken
REFRESH_SECRET=tesToken
// following is optional and should be set while using google OAuth. set some random string otherwise
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```
- and then use
    - `make run`
- to build
    - `make build VERSION=1.0.0`