import pandas as pd
import json
import matplotlib.pyplot as plt

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

    # --- Analysis ---

    # Extract strategy for each player in each game
    player_joins = df[df['event_type'] == 'player_join'].copy()
    if player_joins.empty:
        print("\nNo 'player_join' events found. Cannot analyze by strategy.")
        return

    player_joins_payload = pd.json_normalize(player_joins['payload'])
    game_strategies = pd.concat([
        player_joins[['game_id']].reset_index(drop=True),
        player_joins_payload[['player_id', 'strategy']].reset_index(drop=True)
    ], axis=1)

    # For now, we'll just take the strategy of the first player for each game.
    # This is a simplification that works for single-player games.
    game_strategies = game_strategies.drop_duplicates(subset='game_id', keep='first')


    # Analyze game outcomes
    game_over_df = df[df['event_type'] == 'game_over'].copy()
    if not game_over_df.empty:
        payload_df = pd.json_normalize(game_over_df['payload'])
        game_over_df = pd.concat([game_over_df.reset_index(drop=True), payload_df], axis=1)

        if 'winner' not in game_over_df.columns:
            game_over_df['winner'] = ''
        game_over_df['result'] = game_over_df['winner'].apply(lambda x: 'win' if pd.notna(x) and x != '' else 'loss')

        # Merge strategy information
        game_results = pd.merge(game_over_df, game_strategies, on='game_id')

        print("\n--- Win Rate per Strategy ---")
        if not game_results.empty:
            win_rate = game_results.groupby('strategy')['result'].apply(lambda x: (x == 'win').mean()).reset_index(name='win_rate')
            print(win_rate)
        else:
            print("No completed games to analyze for win rate.")


    # Calculate Game Progress Score
    cards_played = df[df['event_type'] == 'play_card'].groupby('game_id').size().reset_index(name='cards_played')
    if not cards_played.empty:
        progress_df = pd.merge(cards_played, game_strategies, on='game_id')

        # For winning games, progress is 100% (98 cards)
        if not game_over_df.empty and 'game_results' in locals():
            winning_games = game_results[game_results['result'] == 'win']['game_id']
            progress_df.loc[progress_df['game_id'].isin(winning_games), 'cards_played'] = 98

        progress_df['progress_score'] = (progress_df['cards_played'] / 98.0) * 100

        print("\n--- Game Progress Score per Strategy ---")
        avg_progress = progress_df.groupby('strategy')['progress_score'].mean().reset_index(name='average_progress_score')
        print(avg_progress)
    else:
        print("\nNo cards played in any game.")


if __name__ == '__main__':
    main() 