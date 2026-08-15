CREATE TABLE IF NOT EXISTS followers(
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, follower_id)
);

CREATE INDEX IF NOT EXISTS followers_user_idx ON followers(user_id);
CREATE INDEX IF NOT EXISTS followers_follower_idx ON followers(follower_id);