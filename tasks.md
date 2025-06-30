# Implementation Tasks for The Game

This document outlines the development tasks required to build "The Game" according to the architecture specified in `design.md`.

## Phase 1: Core Backend Setup

- [x] **1. Project Dependencies & Tooling:**

  - [x] Install Docker and Docker Compose for managing local services.
  - [x] Add Go dependencies for Redis (`go-redis/redis`) and PostgreSQL (`jackc/pgx`).
  - [x] Re-introduce a `Makefile` to simplify `protoc` generation and other common commands.

- [x] **2. Protobuf and Gateway Generation:**

  - [x] Update `proto/game.proto` with all required RPCs (`CreateGame`, `JoinGame`, `PlayCard`, `StreamGameState`) and their HTTP annotations for the gRPC-Gateway.
  - [x] Ensure all necessary `protoc` plugins (`protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`) are correctly installed and in the `PATH`.
  - [x] Generate the gRPC Go code, gRPC-Gateway code, and OpenAPI definitions.

- [x] **3. Service Setup (Docker):**

  - [x] Create a `docker-compose.yml` file to run Redis and PostgreSQL containers.
  - [x] Define a basic PostgreSQL schema in an `init.sql` script to create the necessary tables (`games`, `moves`, `players`, etc.).

## Phase 2: Server Logic & Database Integration

- [x] **1. Database Layer:**

  - [x] Create a new `pkg/storage` package.
  - [x] Implement connection logic for both Redis and PostgreSQL in `pkg/storage`.
  - [x] Define functions for all database interactions (e.g., `CreateGame`, `GetGameState`, `UpdateGameState`, `SaveMove`).
  - [x] Test database connections with seed data.

- [x] **2. Game Server Implementation:**

  - [x] Update `pkg/server/server.go` to implement the full `GameService` gRPC interface.
  - [x] **CreateGame:**
    - [x] Generate a unique game ID.
    - [x] Initialize the game state (deck, piles) and store it in Redis.
    - [x] Create a corresponding entry in the PostgreSQL `games` table.
  - [x] **JoinGame:**
    - [x] Add a player to an existing game.
  - [x] **PlayCard:**
    - [x] Fetch game state from Redis.
    - [x] Validate the move against the game rules.
    - [x] Update the game state in Redis.
    - [x] Send a response back to the client immediately.
    - [x] Asynchronously write the move details to the PostgreSQL `moves` table.
  - [x] **StreamGameState:**
    - [x] Implement a mechanism (e.g., Redis Pub/Sub) to publish game state changes.
    - [x] Subscribe to these changes and stream them to connected clients.

- [x] **3. Unit Tests for Server Logic:**
  - [x] Set up a test suite for the `server` package with a real storage backend.
  - [x] **Test CreateGame:**
    - [x] Verify that a new game is created in both PostgreSQL and Redis.
    - [x] Check that the initial game state (deck, piles, hand) is correct.
  - [x] **Test JoinGame:**
    - [x] Verify that a second player can join a game.
    - [x] Check that the new player is dealt a hand and the deck size decreases.
  - [x] **Test PlayCard:**
    - [x] Test a valid card play and verify the state update.
    - [x] Test an invalid card play (e.g., wrong value for a pile) and verify the error.
    - [x] Test the "10-back" rule.
  - [x] **Test StreamGameState:**
    - [x] Verify that a client receives the initial state upon connecting.

## Phase 3: Application Assembly & Finalization

- [x] Test `CreateGame`
- [x] Test `JoinGame`
- [x] Test `PlayCard`
- [x] Test `StreamGameState`
- [x] Implement `Storer` interface and mock for unit testing
  - [x] `CreateGame`
  - [x] `JoinGame`
  - [x] `PlayCard`
- [x] `StreamGameState`
- [x] Move existing tests to `_integration_test.go`
- [x] Write true unit tests in `_test.go`
- [x] Complete Integration Tests
  - [x] Test `CreateGame`
  - [x] Test `JoinGame`
  - [x] Test `PlayCard`
  - [x] Test `StreamGameState`

## Phase 3: Assemble Application

- [x] Update `cmd/server/main.go` to use the new `Server` and `Store`
- [x] Update `cmd/client/main.go` to be a simple CLI to interact with the server
- [x] Run the server and client
- [x] Final verification and cleanup

## Phase 4: Documentation and Cleanup

- [x] Review and update `README.md`
- [x] Review and update `design.md`
- [x] Ensure all code is formatted and linted
- [x] Final commit and push
