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

Full usage details can be found in the [Postman collection](collections/postman_collection.json) & [Swagger](collections/swagger.yaml).

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

## Can It Handle Millions of Users? Phhhh, take my beer
Let me check. So what is 1m of users? 

##### 1. Each user in DB takes:
  - `id` = 8 bytes
  - `username` = 50 bytes 
  - `timestamp` = 8 bytes

  Total: 66 bytes

  Total per 1m: 66 * 1_000_000 = 66_000_000 bytes = 0.066 GB seems enough

##### 2. Imagine that each user is subscribed on each: 

  Each subscription it's:
  - 2 `id` = 8 bytes
  - `timestamp` = 8 bytes

  Total: 16 bytes

  Total per 1m: 1_000_000 * 999_999 * 16 = 15_999_984_000_000 bytes = 15.99998 TB -- it looks crazy actually, but it's the worst situation

  We can check the [statistic](https://www.pewresearch.org/internet/2019/04/24/sizing-up-twitter-users/) based on real twitter usage is way more less. 
  In link before, average user has 74 followings and top 10% -- 456, that seems less. So, let's try to use this data:

  Total per 1m: 1_000_000 * 456 * 16 = 7_296_000_000 bytes = 7.296 GB and that's better! 

##### 3. Imagine that everyone will post 1 tweet:

  Each tweet is:
  - 2 `id` = 8 bytes
  - `timestamp` = 8 bytes
  - 280 characters = 280 bytes

  Total 296 bytes

  Total per 1m: 1_000_000 * 296 = 296_000_000 bytes = 0.296 GB

  So it looks like, it perfectly suits even the lowcost server with redis (if we store everything in redis cache)

##### 4. And there is a redis system that has to handle everything:

  - `followers:<id>`. It's a list of followers, so it's the same as in db, we have to store 1_000_000 * 456 * 8 = 3_648_000_000 = 3.648 GB (because of the ids) of data
  - `timeline:<id>`. It's more interesting, if we have 1m active users, it's 1m of this lists, but they are fixed by `max_tweets_timeline_items` in config and by default is 10, so there is 10 tweets id per 8 byte (cz it's a int64) & in total it's 10 * 8 * 1_000_000 = 80_000_000 bytes = 0.08 GB. According to previous statistics average user post 1 tweet per month, top 10% 150 per month. It's nearly 3 tweets per day. 
  - `tweets:global`. It's a list of global tweets and it's also fixed by `max_tweets_to_keep` and by default it's 1000 & as it a list of id: 1000 * 8 = 8_000 bytes = 8e-6 (it's veeeery small number) GB

  And in the end, as all tweets a kept in redis (as a cache with inspiration time, not all the time), it's gonna be 0.296 GB of data

  So we see, that our enemy are user followings and it's a place of discussions.

  At all, we need nearly 4-5 GB ram, it seems pretty small. 


All in all the system can handle 1_000_000 of users and we wouldn't be bankrupted by DigitalOcean or other cloud provider:

1. The Premium intel droplet costs ~100usd (16GB ram + 4 Intel CPU + 8TB transfer)
2. Database with 100GB costs: 10 usd per month (in worth case it's 17000 usd, but here it's better to buy a hardware server razer than VPS)

Total price is: 110 usd per month and looks reasonable for a 1m users platform, isn't it?

But yes, it's too optimistic. In real world we would need more than 4 CPU server, cz fx Redis is one thread (as far as I remember) & to handle 1m users in 1 thread seems impossible. Additionally, the database has to have indexes and it can add 70-150% of overflow, so it will not be actually 8GB of data, but 16 Gb minimum. And Redis also keeps keys for everything and we will have a not significant growth.

I think that finally we have to be ready to pay 500-1000 usd per server to keep 1m users. But it's still pretty good. 

### WHY is it good?

All thees users can generate a batch of money. So what can be done:

1. We can sell users data: 
  - where the tweet is post
  - what phone is used
  - what version of system
  - when tweet was post
2. We can put advertisement to the platform & depending on provider it can bring lots of money. If we have 1m of active users & the normal conversion is 2-3%, we have 20_000 - 30_000 users that "clicks" the add, and per each we have ~0.001usd revenue: 20_000 * 0.001 = 20 - 30 usd. Seems not too much, but depends on the provider. Additionally, it's only clicks, but posting the add can be more profitable.
3. We can bring new features for users with "Premium" subscription. And here even 1% brings a lot: 10_000 * 3usd = 30_000 usd, so that can keep our servers working. 

Sure, we have to include taxes, developing costs & management. But at least even this can bring some real revenue.
