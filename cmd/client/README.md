# The Game - Client CLI

This directory contains the source code for a simple command-line interface (CLI) client to interact with the `the_game` server.

## Prerequisites

Before using the client, ensure the main server is running:

```bash
# From the project root
go build -o bin/server ./cmd/server/main.go
./bin/server
```

Also, make sure the database containers are running via Docker Compose:

```bash
# From the project root
docker-compose up -d
```

## Building the Client

From the project root, you can build the client executable:

```bash
go build -o bin/client ./cmd/client/main.go
```

## Usage

The client uses subcommands to interact with the server's gRPC endpoints.

### 1. `create`

Creates a new game.

**Usage:**

```bash
./bin/client create --player=<your_player_id>
```

**Example:**

```bash
./bin/client create --player="PlayerOne"
```

This will return the new game's state, including the unique `game_id` which you will need for other commands.

### 2. `join`

Joins an existing game.

**Usage:**

```bash
./bin/client join --game=<game_id> --player=<your_player_id>
```

**Example:**

```bash
./bin/client join --game="425de144-a5e3-4d08-a238-28a6938caca9" --player="PlayerTwo"
```

### 3. `play`

Plays a card from your hand onto a pile.

**Usage:**

```bash
./bin/client play --game=<game_id> --player=<your_player_id> --card=<card_value> --pile=<pile_id>
```

- `pile_id` can be one of: `up1`, `up2`, `down1`, `down2`.

**Example:**

```bash
./bin/client play --game="425de144-a5e3-4d08-a238-28a6938caca9" --player="PlayerOne" --card=7 --pile="up1"
```

### 4. `stream`

Connects to a game and streams its state in real-time. Any updates (like another player joining or playing a card) will be printed to the console. This command will run until you manually stop it (e.g., with `Ctrl+C`).

**Usage:**

```bash
./bin/client stream --game=<game_id>
```

**Example:**

```bash
./bin/client stream --game="425de144-a5e3-4d08-a238-28a6938caca9"
```
