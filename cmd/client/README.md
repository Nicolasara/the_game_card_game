# The Game - Client CLI

This directory contains the source code for a command-line interface (CLI) client to interact with `the_game` server. This client features a Terminal User Interface (TUI) to provide a rich, visual representation of the game state.

## Prerequisites

Before using the client, ensure the main server is running:

```bash
# From the project root, first build the server
go build -o bin/server ./cmd/server/main.go

# Then run it
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

Creates a new game and displays its state in an interactive TUI.

**Usage:**

```bash
./bin/client create --player=<your_player_id>
```

This command will take over the terminal to display the game board. Press `q` or `Ctrl+C` to quit and return to the command line. The `game_id` will be printed to the logs when the server creates the game.

### 2. `join`

Joins an existing game and displays its state in the TUI.

**Usage:**

```bash
./bin/client join --game=<game_id> --player=<your_player_id>
```

Like `create`, this will display the game board. Press `q` or `Ctrl+C` to quit.

### 3. `play`

Plays a card from your hand onto a pile. This is a one-shot command that prints a success or failure message and then exits. It does _not_ launch the TUI. To see the results of your play, use the `stream` command.

**Usage:**

```bash
./bin/client play --game=<game_id> --player=<your_player_id> --card=<card_value> --pile=<pile_id>
```

- `pile_id` can be one of: `up1`, `up2`, `down1`, `down2`.

**Example:**

```bash
./bin/client play --game="ff05981f-..." --player="PlayerOne" --card=8 --pile="up2"
```

### 4. `stream`

Connects to a game and streams its state in real-time. The TUI will automatically update whenever another player makes a move. This is the best way to get a live view of the game.

**Usage:**

```bash
./bin/client stream --game=<game_id> --player=<your_player_id>
```

The application will run until you manually stop it by pressing `q` or `Ctrl+C`.
