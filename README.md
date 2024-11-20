# go-adjoe-task
**how to use**

- To start, please install: docker, docker-compose and go

 - `'make up'` on root folder run it will start test-task docker
   container with all requied dependencies. 
  - `'make bash'` to open container bash window 
  - `localhost:3333` to access go container from localhost
  - `make aws-cli foo bar` to execute the aws cli with parameters foo bar
  - `make aws-cli sqs list-queues` to access the sqs queue on aws localstack
  - If you import dependencies to your go code, please use `make stop` and `make up` again to automatically download them
  - To access the sql server, use mysql:3306 as address and Port.
  - In your go application you have to use the eu-central-1 region to access aws localstack
- In localstack, run `aws configure` and set `test` as aws access id and aws secret. It will allow sqs queue to connect from go client
- To start the program, please hit the heartbeart URL - [http://localhost:3333/](http://localhost:3333/)

## URLs

- curl --location 'http://localhost:3333/add/a/7/b/5/'
- curl --location 'http://localhost:3333'
- curl --location 'http://localhost:3333/fetch/sum'
- curl --location 'http://localhost:3333/sqs/message'
```text
curl --location 'http://localhost:3333/sqs/sendmessage' \
--header 'Content-Type: application/json' \
--data '{
"message": "Second message 2"
}'
```
