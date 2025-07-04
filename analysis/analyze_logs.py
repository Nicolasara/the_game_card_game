import pandas as pd
import json
import matplotlib.pyplot as plt
import seaborn as sns

def load_logs(file_path='logs/game_logs.jsonl'):
    """
    Load game logs from a JSONL file into a pandas DataFrame.
    """
    records = []
    with open(file_path, 'r') as f:
        for line in f:
            try:
                records.append(json.loads(line))
            except json.JSONDecodeError:
                print(f"Skipping malformed line: {line.strip()}")
    return pd.DataFrame(records)

def analyze_single_player_performance(df):
    """
    Analyzes and plots performance for single-player games.
    """
    print("\n--- Analyzing Single-Player Game Performance ---")

    # Determine number of players per game by counting unique player_ids in 'player_join' events
    player_joins = df[df['event_type'] == 'player_join'].copy()
    if player_joins.empty:
        print("\nNo 'player_join' events found. Cannot determine player counts or strategies.")
        return

    player_joins_payload = pd.json_normalize(player_joins['payload'])
    player_joins = pd.concat([
        player_joins[['game_id']].reset_index(drop=True),
        player_joins_payload[['player_id', 'strategy']].reset_index(drop=True)
    ], axis=1)

    game_player_counts = player_joins.groupby('game_id')['player_id'].nunique().reset_index(name='num_players')

    single_player_games = game_player_counts[game_player_counts['num_players'] == 1]
    sp_game_ids = single_player_games['game_id']

    if sp_game_ids.empty:
        print("No single-player games found in logs.")
        return

    print(f"Found {len(sp_game_ids)} single-player games.")
    sp_df = df[df['game_id'].isin(sp_game_ids)]

    # Extract strategy for each player (for single player games, this is straightforward)
    game_strategies = player_joins[player_joins['game_id'].isin(sp_game_ids)][['game_id', 'strategy']]

    # --- Analyze game outcomes ---
    game_over_df = sp_df[sp_df['event_type'] == 'game_over'].copy()
    if not game_over_df.empty:
        # Determine win/loss by counting played cards. A win requires all 98 cards to be played.
        cards_played_count = sp_df[sp_df['event_type'] == 'play_card'].groupby('game_id').size()
        winning_games = cards_played_count[cards_played_count == 98].index

        game_over_df['result'] = game_over_df['game_id'].apply(lambda id: 'win' if id in winning_games else 'loss')
        game_results = pd.merge(game_over_df, game_strategies, on='game_id')

        print("\n--- Win Rate per Strategy (Single-Player) ---")
        win_rate = game_results.groupby('strategy')['result'].apply(lambda x: (x == 'win').mean()).reset_index(name='win_rate')
        win_rate['win_rate'] = win_rate['win_rate'] * 100
        print(win_rate)

        # Plotting Win Rate
        plt.figure(figsize=(10, 6))
        sns.barplot(data=win_rate, x='strategy', y='win_rate')
        plt.title('Win Rate per Strategy (Single-Player Games)')
        plt.ylabel('Win Rate (%)')
        plt.xlabel('Strategy')
        plt.xticks(rotation=45)
        plt.tight_layout()
        plt.savefig('analysis/win_rate_single_player.png')
        print("Saved win rate plot to analysis/win_rate_single_player.png")

    # --- Calculate Game Progress Score ---
    cards_played = sp_df[sp_df['event_type'] == 'play_card'].groupby('game_id').size().reset_index(name='cards_played')
    if not cards_played.empty:
        progress_df = pd.merge(cards_played, game_strategies, on='game_id')

        if not game_over_df.empty:
            # Get the winning games from the game_over_df directly
            winning_games = game_over_df[game_over_df['result'] == 'win']['game_id']
            progress_df.loc[progress_df['game_id'].isin(winning_games), 'cards_played'] = 98

        progress_df['progress_score'] = (progress_df['cards_played'] / 98.0) * 100

        print("\n--- Game Progress Score per Strategy (Single-Player) ---")
        avg_progress = progress_df.groupby('strategy')['progress_score'].mean().reset_index(name='average_progress_score')
        print(avg_progress)

        # Plotting Game Progress
        plt.figure(figsize=(10, 6))
        sns.barplot(data=avg_progress, x='strategy', y='average_progress_score')
        plt.title('Average Game Progress Score per Strategy (Single-Player Games)')
        plt.ylabel('Progress Score (%)')
        plt.xlabel('Strategy')
        plt.xticks(rotation=45)
        plt.tight_layout()
        plt.savefig('analysis/progress_score_single_player.png')
        print("Saved progress score plot to analysis/progress_score_single_player.png")


def main():
    """
    Main function to load, analyze, and visualize game logs.
    """
    df = load_logs()

    if df.empty:
        print("Log file is empty. No analysis to perform.")
        return

    print("Log data loaded successfully.")
    print(f"Total number of events logged: {len(df)}")
    print("\nEvent types distribution:")
    print(df['event_type'].value_counts())

    analyze_single_player_performance(df)


if __name__ == '__main__':
    main() 