# Implementation Tasks for The Game

This document outlines the development tasks required to build "The Game" according to the architecture specified in `design.md`.

## Phase 1: Core Backend Setup

- [ ] **1. Project Dependencies & Tooling:**

  - [ ] Install Docker and Docker Compose for managing local services.
  - [ ] Add Go dependencies for Redis (`go-redis/redis`) and PostgreSQL (`jackc/pgx`).
  - [ ] Re-introduce a `Makefile` to simplify `protoc` generation and other common commands.

- [ ] **2. Protobuf and Gateway Generation:**

  - [ ] Update `proto/game.proto` with all required RPCs (`CreateGame`, `JoinGame`, `PlayCard`, `StreamGameState`) and their HTTP annotations for the gRPC-Gateway.
  - [ ] Ensure all necessary `protoc` plugins (`protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`) are correctly installed and in the `PATH`.
  - [ ] Generate the gRPC Go code, gRPC-Gateway code, and OpenAPI definitions.

- [ ] **3. Service Setup (Docker):**
  - [ ] Create a `docker-compose.yml` file to run Redis and PostgreSQL containers.
  - [ ] Define a basic PostgreSQL schema in an `init.sql` script to create the necessary tables (`games`, `moves`, `players`, etc.).

## Phase 2: Server Logic & Database Integration

- [ ] **1. Database Layer:**

  - [ ] Create a new `pkg/storage` package.
  - [ ] Implement connection logic for both Redis and PostgreSQL in `pkg/storage`.
  - [ ] Define functions for all database interactions (e.g., `CreateGame`, `GetGameState`, `UpdateGameState`, `SaveMove`).

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
