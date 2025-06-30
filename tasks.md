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

- [ ] **3. Service Setup (Docker):**
  - [ ] Create a `docker-compose.yml` file to run Redis and PostgreSQL containers.
