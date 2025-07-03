# Bot-Based Game Strategy Analysis: Design Document

## 1. Overview

This document outlines the design for a bot-driven system to analyze and determine optimal strategies for "The Game". The goal is to create multiple bot clients, each implementing a different strategy, have them play the game thousands of times, and collect data to identify the most effective approaches.

## 2. Key Components

The project will be broken down into the following key components:

- **Bot Client**: A new command-line application that can play the game autonomously. It will connect to the existing game server and make moves based on a predefined strategy.
- **Strategy Interface**: A well-defined Go interface that all bot strategies will implement. This will allow for easy integration of new strategies.
- **Initial Strategies**: A set of simple, baseline strategies to begin with:
  - **Random**: Plays any valid card on any valid pile. This will serve as our performance baseline.
  - **Minimal Jump**: A "greedy" strategy that always chooses the valid move with the smallest difference between the card in hand and the top card of the pile.
  - **Safe Ten**: A defensive strategy that prioritizes playing a card that is exactly 10 different from the pile, if possible.
- **Game Data Collection**: The game server will be enhanced to log detailed game events to a structured format.
- **Data Analysis Framework**: A suite of scripts or tools to process the collected game data and generate insightful statistics.

## 3. Design Decisions

### 3.1. Bot Client Implementation

The bot client will be a new Go application, likely under `cmd/bot/main.go`. It will leverage the existing protobuf client and game logic. The bot will be configurable to run a specific strategy.

### 3.2. Strategy Interface

A `Strategy` interface will be defined in a new `pkg/bot` package.

```go
package bot

import (
	pb "the_game_card_game/proto"
)

// Strategy defines the interface for a game-playing bot.
type Strategy interface {
	// GetNextMove determines the next move for the bot to make.
	GetNextMove(gameState *pb.GameState) (*pb.PlayCardRequest, error)
}
```

### 3.3. Data Collection

To capture the necessary data for analysis, we will implement a structured logging system within the game server.

- **Format**: Game events will be logged as single-line JSON objects to a dedicated log file (e.g., `game_logs.jsonl`). This format is easy to parse line-by-line.
- **Game Identification**: Each game will be assigned a unique `game_id`. This UUID will be included in every log entry for that game, allowing us to easily group all related events.
- **Event-Driven Logging**: We will log discrete events rather than just state snapshots. Key events to log include:
  - `game_start`: Contains the initial configuration for the game, including `num_players` and an array of the `player_strategies` being used.
  - `play_card`: Logs the card played, the pile it was played on, the player's identity, and their strategy.
  - `end_turn`: Captures the state at the end of a player's turn.
  - `game_over`: Records the final outcome (win/loss), the reason, and summary statistics like total cards played and cards remaining.

### 3.4. Data Analysis

We will start with simple analysis scripts in a language like Python with libraries such as Pandas and Matplotlib. These scripts will calculate the metrics defined in the following section.

### 3.5. Metrics and Analysis

To determine the effectiveness of different strategies, we will focus on the following metrics and dimensions of analysis.

#### 3.5.1. Defining the "Best" Strategy

A strategy's success will be measured by a combination of metrics:

1.  **Win Rate**: The primary metric. The percentage of games won out of all games played with a given strategy.
2.  **Game Progress Score (for losses)**: When a game is lost, we still want to know how "well" the strategy performed. This will be calculated as `(98 - cards_remaining) / 98`. This normalizes the number of cards successfully played onto the board.

#### 3.5.2. Dimensions of Analysis

We will analyze the performance of strategies across several dimensions:

- **Strategy vs. Strategy**: Direct comparison of the metrics above for different strategies.
- **Performance by Player Count**: How does a strategy's effectiveness change with 1, 2, 3, or 4 players?
- **Performance by Opponent Mix**: How does a strategy perform when paired with other specific strategies (e.g., how does a "safe" strategy do when playing with "greedy" players)?

#### 3.5.3. Additional Proposed Metrics

To gain deeper insights, we will also track:

- **Game Duration**: The number of turns a game lasts. Do better strategies lead to shorter or longer games?
- **Pile Health**: The average difference between cards played on a pile. Lower differences might indicate a more controlled, sustainable strategy.
- **"Backwards-10" Play Frequency**: How often is the special rule of playing a card exactly 10 away used? This could be a sign of a flexible strategy or a desperate move.

## 4. Scalability

To run a large number of simulations, we will need to run multiple bot clients in parallel. The Docker-based setup will be extended to support launching multiple bot instances against a single game server.

---
