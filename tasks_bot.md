# Task List: Bot-Based Game Strategy Analysis

## Phase 1: Foundational Setup

- [ ] Create `cmd/bot` directory for the new bot client.
- [ ] Implement the basic bot client that can connect to the game server.
- [ ] Define the `Strategy` interface in a new `pkg/bot` package.
- [ ] Create a simple "random play" strategy as the first implementation.

## Phase 2: Data Collection

- [ ] Enhance the game server to log game events to a file.
- [ ] Decide on a structured logging format (JSON is a good starting point).
- [ ] Ensure all relevant game actions (play card, end turn, game over) are logged.
- [ ] Implement structured JSON logging in the server to a `game_logs.jsonl` file.
- [ ] Log the `game_start` event with `game_id`, `num_players`, and `player_strategies`.
- [ ] Log `play_card`, `end_turn`, and `game_over` events with all relevant context.

## Phase 3: Initial Strategies

- [ ] Implement a "Random" strategy for a baseline comparison.
- [ ] Implement a "Minimal Jump" (greedy) strategy.
- [ ] Implement a "Safe Ten" (defensive) strategy.

## Phase 4: Analysis and Visualization

- [ ] Set up a Python environment with Pandas and Matplotlib.
- [ ] Write a script to parse the game log file.
- [ ] Calculate and print basic statistics (win rate, game duration) for each strategy.
- [ ] Create simple plots to visualize the performance of different strategies.
- [ ] Set up a Python analysis environment (e.g., with Pandas, Matplotlib, Jupyter).
- [ ] Write a script to parse the `game_logs.jsonl` file into a Pandas DataFrame.
- [ ] Calculate primary metrics: Win Rate and Game Progress Score.
- [ ] Analyze and plot performance based on player count and opponent strategy mix.
- [ ] Analyze and plot secondary metrics: Game Duration, Pile Health, "Backwards-10" frequency.

## Phase 5: Scaling and Refinement

- [ ] Update `docker-compose.yml` to allow running multiple bot clients in parallel.
- [ ] Run a large number of simulations (e.g., 10,000 games) for each strategy.
- [ ] Refine the data analysis scripts to provide deeper insights.
- [ ] Document the findings and the performance of each strategy.
