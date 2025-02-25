# Event processing challenge

A casino has a few different types of events coming in that we would like to
process to get an insight into casino performance and player activity.

For this assignment, you will need to implement a set of components that:

1. publish generated events,
1. subscribe to these events,
1. enrich them using data from various sources (HTTP API, database, and in-memory mapping),
1. materialize various aggregates,
1. output events as logs.

## Setup

Clone the repository and run `make`. This will start the `database` service, which is a Postgres server, and run the needed DB migrations.

Optionally, you can run `make generator` to see how the generator works. It will run for 5 seconds, logging the generated events, then exit.

You are now ready to start building the components.

## Event structure

You can find the event structure in [./internal/casino/event.go](./internal/casino/event.go).

## Publish

A `generator` service has been provided that generates random events. Add the ability to publish these events. Feel free to decide which transport and/or technology to use.

The `publisher.go` file contains the logic to publish generated events to RabbitMQ. It connects to RabbitMQ, declares a queue named `casino_events`, serializes the event to JSON, and publishes it to the queue. Here is a brief description of what happens in `publisher.go`:

- Connect to RabbitMQ using the URL from the configuration.
- Declare a queue named `casino_events`.
- Serialize the event to JSON format.
- Publish the event to the `casino_events` queue.

In `main.go`, the `publishGeneratedEvents` function is responsible for publishing generated events to RabbitMQ using the `PublishEvent` function from `publisher.go`. Here is how it is used in `main.go`:

- The `publishGeneratedEvents` function is called in a separate goroutine to publish events to RabbitMQ.


## Subscribe

Implement a service that will subscribe to the published events. Again, feel free to decide which technology to use.

The `subscriber.go` file contains the logic to subscribe to events from RabbitMQ. It connects to RabbitMQ, declares a queue named `casino_events`, and listens for messages on the queue. When a message is received, it processes the event and updates the materialized stats. Here is a brief description of what happens in `subscriber.go`:

- Connect to RabbitMQ using the URL from the configuration.
- Declare a queue named `casino_events`.
- Listen for messages on the `casino_events` queue.
- Process the event and update the materialized stats.

In `main.go`, the `subscribeToProcessedEvents` function is responsible for subscribing to events from RabbitMQ using the `SubscribeEvents` function from `subscriber.go`. Here is how it is used in `main.go`:

- The `subscribeToProcessedEvents` function is called in a separate goroutine to subscribe to events from RabbitMQ and update the materialized stats.

### Why RabbitMQ?

We use RabbitMQ for our pub/sub system due to its reliability, ease of use. RabbitMQ provides robust messaging capabilities, including message durability, acknowledgments, and flexible routing. It is well-suited for scenarios where message delivery guarantees and complex routing are required.

### Alternative: Kafka

While we could use Kafka, which is known for its high throughput and scalability, we chose RabbitMQ for this project because it is simpler to set up and manage for our current needs. Kafka is more suitable for large-scale, distributed systems with high message volumes, but RabbitMQ provides sufficient performance and features for our use case.

## Enrich

Implement 3 components (services) that will receive the events and enrich them:

### Common currency

To be able to analyze `bet` and `deposit` events, we need them in a common currency.

If `Currency` is not `EUR`, we need to convert the `Amount` to `EUR`. For this, use API endpoint https://api.exchangerate.host/latest. Store the `EUR` amount into `AmountEUR` field.

If `Currency` is already `EUR`, set `AmountEUR` to the same value as `Amount`.

API results may be cached for up to 1 minute. Feel free to decide on the kind of caching technology you want to use.

The `currency.go` file contains the logic to convert different currencies to EUR. It fetches exchange rates from an external API and performs the conversion. Here is a brief description of what happens in `currency.go`:

- Fetch exchange rates from the API.
- Cache the exchange rates for up to 1 minute.
- Convert the event amount to EUR using the fetched exchange rates.

### Player data

We have a Postgres database (`database` service) where you can find a `players` table with some of the players inserted.

We need to look up a player for each event and store their data into `Player` field. If there is no data for a player, log that it's missing and leave the field as zero-value.

DB results may not be cached.

The `player.go` file contains the logic to fetch player information from the database. It uses a repository pattern to handle database operations. Here is a brief description of what happens in `player.go`:

- Connect to the Postgres database.
- Fetch player information using the player ID from the event.
- Return the player information or log a warning if the player is not found.

### Human-friendly description

We need to represent each event with a human-friendly description. Examples:

```json
{
  "id": 1,
  "player_id": 10,
  "game_id": 100,
  "type": "game_start",
  "created_at": "2022-01-10T12:34:56.789+00"
}
```

```
Player #10 started playing a game "Rocket Dice" on January 10th, 2022 at 12:34 UTC.
```

```json
{
  "id": 2,
  "player_id": 11,
  "game_id": 101,
  "type": "bet",
  "amount": 500,
  "currency": "USD",
  "amount_eur": 468,
  "created_at": "2022-02-02T23:45:67.89+00",
  "player": {
    "email": "john@example.com",
    "last_signed_in_at": "2022-02-02T23:01:02.03+00"
  }
}
```

```
Player #11 (john@example.com) placed a bet of 5 USD (4.68 EUR) on a game "It's bananas!" on February 2nd, 2022 at 23:45 UTC.
```

```json
{
  "id": 3,
  "player_id": 12,
  "type": "deposit",
  "amount": 10000,
  "currency": "EUR",
  "created_at": "2022-03-03T12:12:12+00"
}
```

```
Player #12 made a deposit of 100 EUR on February 3rd, 2022 at 12:12 UTC.
```

You can find the mapping of game titles in [./internal/casino/game.go](./internal/casino/game.go).

The `description.go` file contains the logic to generate human-readable descriptions for events. It formats the event details into a string. Here is a brief description of what happens in `description.go`:

- Format the event timestamp.
- Retrieve the game title using the game ID from the event.
- Generate a description based on the event type and details.

### Usage in `main.go`

In `main.go`, the enrichment components are used to process events before publishing them to RabbitMQ. Here is how they are used:

- The `publishGeneratedEvents` function fetches player data using `player.go`, converts the event amount to EUR using `currency.go`, and generates a human-friendly description using `description.go`.
- The enriched event is then published to RabbitMQ.

## Materialize

We would like to materialize the following data:

- total number of events,
- number of events per minute,
- events per second as a moving average in the last minute,
- top player by number of bets,
- top player by number of wins,
- top player by sum of deposits in EUR.

Feel free to decide on the algorithm, technology or library you want to use. This data should not be persisted and should only be available while the components are running.

Data should be available via an HTTP API in the following form:

```
GET http://localhost/materialized
```

```json
{
  "events_total": 12345,
  "events_per_minute": 123.45,
  "events_per_second_moving_average": 3.12,
  "top_player_bets": {
    "id": 10,
    "count": 150
  },
  "top_player_wins": {
    "id": 11,
    "count": 50
  },
  "top_player_deposits": {
    "id": 12,
    "count": 15000
  }
}
```

- Track the total number of events.
- Calculate the number of events per minute.
- Calculate the events per second as a moving average over the last minute.
- Track the top player by number of bets, wins, and sum of deposits in EUR.
- Provide an HTTP API to retrieve the materialized data.

### Algorithm, Technology, or Library

We use Go's standard library for HTTP server functionality and synchronization primitives. The algorithm involves maintaining in-memory counters and maps to track event statistics. We use a mutex to ensure thread-safe updates to these statistics. The moving average is calculated by maintaining a sliding window of event timestamps.

### Usage in `main.go`

In `main.go`, the materialization component is used to update and serve event statistics. Here is how it is used:

- The `subscribeToProcessedEvents` function updates the materialized data using `materialize.go`.
- The `StartHTTPServer` function in `materialize.go` starts an HTTP server to serve the materialized data.

## Output

Log the events in their final form to standard output. Logs should be in JSON format and use the same keys as the `Event` type.

## Updates in `docker-compose.yml`

`docker-compose.yml` include the following services:

- **generator**: This service runs the Go application that generates events. It uses the `golang:1.17-alpine` image and mounts the current directory to `/app` inside the container. The `go run internal/cmd/generator/main.go` command is executed to start the generator.

- **database**: This service runs a Postgres database using the `postgres:14-alpine` image. It sets the `POSTGRES_USER` and `POSTGRES_PASSWORD` environment variables to `casino`. The database data is stored in the `./db` directory on the host machine and is exposed on port `5432`.

- **rabbitmq**: This service runs RabbitMQ with the management plugin enabled, using the `rabbitmq:3-management` image. It exposes the RabbitMQ service on port `5672` and the management interface on port `15672`.

