# Task List: Bot-Based Game Strategy Analysis

## Phase 1: Foundational Setup

- [x] Create `cmd/bot` directory for the new bot client.
- [x] Implement the basic bot client that can connect to the game server.
- [x] Define the `Strategy` interface in a new `pkg/bot` package.
- [x] Implement a "Random" strategy for a baseline comparison.

## Phase 2: Data Collection

- [x] Implement structured JSON logging in the server to a `game_logs.jsonl` file.
- [x] Log the `game_start` event with `game_id`, `num_players`, and `player_strategies`.
- [x] Log `play_card`, `end_turn`, and `game_over` events with all relevant context.

## Phase 3: Initial Strategies

- [x] Implement a "Minimal Jump" (greedy) strategy.
- [x] Implement a "Safe Ten" (defensive) strategy.
- [x] Implement a "Random" strategy for a baseline comparison.

## Phase 4: Analysis and Visualization

- [x] Set up a Python analysis environment (e.g., with Pandas, Matplotlib, Jupyter).
- [x] Write a script to parse the `game_logs.jsonl` file into a Pandas DataFrame.
- [x] Calculate primary metrics: Win Rate and Game Progress Score.
- [x] Analyze and plot performance based on player count and opponent strategy for a single player game.
- [ ] Analyze and plot performance based on player count and strategy mix for a multi-player game.
- [ ] Analyze and plot secondary metrics: Game Duration, Pile Health, "Backwards-10" frequency.

## Phase 5: Scaling and Refinement

- [x] Update `docker-compose.yml` to allow running multiple bot clients in parallel.
- [x] Run a large number of simulations (e.g., 10,000 games) for each strategy.
- [ ] Refine the data analysis scripts to provide deeper insights.
- [ ] Document the findings and the performance of each strategy.
