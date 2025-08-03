## How to deploy
There are 2 parts of the deployment:
- service part
- application part

### Service Part
The service part is a compose of all services used by the application:
1. Postgres
2. Redis
3. Angie
This services are responded for a well functioning application.

### Application Part
The application part is composed of 4 services:
1. API server
2. Worker
3. Websocket server
4. Migrations: service run database migrations
This part is responsible for the business logic and user interaction.

## How to run
To run the application, you need to have docker installed on your machine.
Then create the external network `twitter-clone-net` using the following command:
```sh
docker network create twitter-clone-net
```
Finally, run the application using the following command:
```sh
docker compose -f docker-compose-services.yml up -d
docker compose -f docker-compose.yml up -d
```

## How to stop
To stop the application, you can use the following command:
```sh
docker compose -f docker-compose.yml down
docker compose -f docker-compose-services.yml down
```

## Security note
To run application in production it's recommended to change all passwords from default ones.

## Images
There is CI job that builds images and pushes them to the registry.
To pull the image, you can use the following commands:
```sh
docker pull ghcr.io/juy13/twitter-clone-api:v0.1
docker pull ghcr.io/juy13/twitter-clone-ws_socket:v0.1
docker pull ghcr.io/juy13/twitter-clone-worker:v0.1
```
