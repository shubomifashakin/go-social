CREATE TABLE IF NOT EXISTS comments(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    content VARCHAR(400) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW() 
);

CREATE INDEX IF NOT EXISTS comments_user_idx ON comments(user_id);
CREATE INDEX IF NOT EXISTS comments_posts_idx ON comments(post_id);