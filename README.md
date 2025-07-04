# The Game

This project is a Go-based, real-time, multiplayer online version of the card game "The Game".

## Game Rules

**Objective:** The goal is to play all 98 cards from the deck into four piles.

**Setup:**

- **Deck:** The game uses a deck of 98 cards, numbered from 2 to 99.
- **Piles:** There are four discard piles on the table.
  - Two piles are **ascending**, starting with the number 1. Cards must be played in increasing order (e.g., on a 10, you can play any card higher than 10).
  - Two piles are **descending**, starting with the number 100. Cards must be played in decreasing order (e.g., on a 90, you can play any card lower than 90).
- **Players:** The game can be played by 1 to 8 players.
- **Hand:** Each player is dealt a hand of cards (the number of cards depends on the player count).

**Gameplay:**

1.  Players take turns.
2.  On each turn, a player must play at least a set number of cards from their hand onto the four piles.
3.  After playing, the player draws cards from the deck to replenish their hand.
4.  The game ends in one of two ways:
    - **Win:** The players win if all 98 cards are successfully played onto the piles.
    - **Loss:** The players lose if a player cannot make a legal move on their turn.

**Special Rule:** A player can play a card that is exactly 10 higher/lower on a descending/ascending pile, respectively, to move the pile's value in the "wrong" direction. For example, if a descending pile is at 87, you can play a 97 on it.

## Running the Project

There are two ways to run this project: locally using `go run`, or with Docker.

### With Docker (Recommended)

This is the easiest way to get the entire application stack running, including the database and Redis.

1.  **Build the Docker images:**

    ```sh
    make docker-build
    ```

2.  **Start all services:**

    ```sh
    make docker-up
    ```

    This will start the game server, PostgreSQL, and Redis in the background.

3.  **Run the client:**
    To create a new game:

    ```sh
    ./bin/client --create --player="YourName"
    ```

    To join an existing game:

    ```sh
    ./bin/client --game <GAME_ID> --player="YourName"
    ```

    _Note: The client binary needs to be built locally first via `make client`._

4.  **Stop all services:**
    ```sh
    make docker-down
    ```

### Local Development

This method is for running the server and client directly on your machine. You will need to have Go, PostgreSQL, and Redis installed.

**Prerequisites:**

- Go (version 1.22 or later)
- `protoc` (Protocol Buffer Compiler)
- PostgreSQL
- Redis

**Running the server:**

```sh
go run ./cmd/server
```

**Running the client:**

```sh
# To create a game
go run ./cmd/client --create --player="YourName"

# To join a game
go run ./cmd/client --game <GAME_ID> --player="YourName"
```
