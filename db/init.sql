-- Create a table to store game sessions
CREATE TABLE IF NOT EXISTS games (
    game_id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    winner VARCHAR(255)
);

-- Create a table to store individual moves within a game
CREATE TABLE IF NOT EXISTS moves (
    move_id SERIAL PRIMARY KEY,
    game_id VARCHAR(255) REFERENCES games(game_id),
    player_id VARCHAR(255) NOT NULL,
    card_played INT NOT NULL,
    pile_id VARCHAR(50) NOT NULL,
    move_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create a table for players (optional, but good for tracking stats)
CREATE TABLE IF NOT EXISTS players (
    player_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
); 