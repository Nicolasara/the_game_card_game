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

- [ ] **2. Game Server Implementation:**
  - [ ] Update `pkg/server/server.go` to implement the full `GameService` gRPC interface.
  - [ ] **CreateGame:**
    - Generate a unique game ID.
    - Initialize the game state (deck, piles) and store it in Redis.
    - Create a corresponding entry in the PostgreSQL `games` table.
  - [ ] **JoinGame:**
    - Add a player to an existing game.
  - [ ] **PlayCard:**
    - Fetch game state from Redis.
    - Validate the move against the game rules.
    - Update the game state in Redis.
    - Send a response back to the client immediately.
    - Asynchronously write the move details to the PostgreSQL `moves` table.
  - [ ] **StreamGameState:**
    - Implement a mechanism (e.g., Redis Pub/Sub) to publish game state changes.
    - Subscribe to these changes and stream them to connected clients.

## Phase 3: Application Assembly & Finalization

- [ ] **1. Main Application (`cmd/server/main.go`):**

  - [ ] Initialize and manage database connections.
  - [ ] Start the gRPC server with the `GameService` implementation.
  - [ ] Configure and run the gRPC-Gateway as a separate goroutine to proxy HTTP requests to the gRPC server.

- [ ] **2. Client & Testing:**

  - [ ] Enhance the `cmd/client/main.go` to test the new RPCs (`JoinGame`, `PlayCard`).
  - [ ] (Optional) Create a simple web-based client using the generated HTTP/JSON endpoints.
  - [ ] Write unit tests for the core game logic.

- [ ] **3. Documentation:**
  - [ ] Update `README.md` with comprehensive instructions on how to set up the environment, run the services using Docker Compose, and interact with the game via both gRPC and HTTP.
