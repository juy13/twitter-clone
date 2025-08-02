# Twitter Clone

## Idea

The goal is to clone Twitter's core functionality:

1. Create a user
2. Follow other users
3. Post tweets
4. Read a timeline

## Architecture

The project is built on a microservices architecture, consisting of:

1. API Service
2. Worker Service
3. WebSocket (WS) Service

### API Service

This service implements the main v1 API endpoints:

* `follow_user`
* `tweet`
* `new_user`
* `followers`
* `followings`
* `get_user`
* `get_tweet`
* `tweets`

Full usage details can be found in the [Postman collection]().

**Dependencies:**

* Redis
* PostgreSQL

This service handles both persistent data storage and the Pub/Sub model.

### Worker Service

This service listens for new tweets and distributes them to users.
It follows a **fan-out model**, allowing followers to receive updates as soon as a tweet is published.

> **IMPORTANT:** A hybrid model is under development and will be available soon.

### WebSocket Service

This service enables users to receive real-time updates to their timelines.
Once a tweet is published and propagated by the worker, it’s delivered to the user (unless the hybrid model is active).

### Redis

Redis is a cornerstone of this architecture and ensures fast tweet distribution and access. It operates with:

**Channels:**

* `tweets:channel`: A newly published tweet, sent to the worker
* `workers:channel`: A processed tweet, sent from the worker to the users

**Lists:**

* `followers:<id>`: List of user IDs who follow a specific user (used for tweet propagation)
* `timeline:<id>`: A user's timeline (tweet IDs)
* `tweets:global`: Global thread of all tweets

**Keys:**

* `tweet:<id>`: Stores tweet content in Redis for quick access (acts as a cache).
  If a tweet is missing in Redis, it falls back to the database.

> **Note 1:** Redis is used here due to its simplicity, but the Pub/Sub logic could be replaced with RabbitMQ or similar tools.
> **Note 2:** Kafka can also be used instead of Redis for more robust queueing and streaming.

### PostgreSQL Database

The primary persistent storage layer. It contains 3 tables:

1. **Users**
2. **Tweets**
3. **Follows**

It serves as the source of truth for all application data.
Although PostgreSQL is used here, any relational database could be used in its place.

## Benefits of This Architecture

### 1. Scalability

The system is designed to scale horizontally:

* API servers can be scaled behind a load balancer
* Redis can be distributed and scaled vertically or horizontally
* Workers can be multiplied across available infrastructure
* WebSocket servers can also be scaled with load balancing

### 2. Component Independence

Each service can be deployed and maintained separately.
They don't need to reside on the same host (though they share core dependencies like Redis and the database).

### 3. Testability

The microservice architecture allows each component to be tested independently, making debugging and testing much easier.

## Can It Handle Millions of Users?
