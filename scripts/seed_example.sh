#!/usr/bin/env bash
set -euo pipefail

# Seed example data for rbg: creates a test user, several feeds, follows and a couple of posts.
# Usage:
#   ./scripts/seed_example.sh
# or set DB_URL environment variable first:
#   DB_URL="postgres://user:pass@host:5432/db?sslmode=disable" ./scripts/seed_example.sh
#
# This script is idempotent: it uses ON CONFLICT DO NOTHING for inserts where appropriate.

DB_URL="${DB_URL-}"

# Try to read DB URL from config file (~/.config/rbg/config.json) using jq if available.
if [ -z "$DB_URL" ] && command -v jq >/dev/null 2>&1 && [ -f "$HOME/.config/rbg/config.json" ]; then
  DB_URL="$(jq -r .db_url ~/.config/rbg/config.json)"
  if [ "$DB_URL" = "null" ]; then
    DB_URL=""
  fi
fi

# Fallback: try to use the CLI to print db-url (go run . db-url or just run db-url)
if [ -z "$DB_URL" ]; then
  if command -v go >/dev/null 2>&1; then
    # attempt `go run . db-url`
    if DB_URL_CANDIDATE="$(go run . db-url 2>/dev/null || true)" && [ -n "$DB_URL_CANDIDATE" ]; then
      DB_URL="$DB_URL_CANDIDATE"
    fi
  fi
fi

if [ -z "$DB_URL" ]; then
  echo "ERROR: couldn't determine DB URL."
  echo "Set DB_URL environment variable or add db_url to ~/.config/rbg/config.json,"
  echo "or ensure 'go run . db-url' prints the DB URL."
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "ERROR: psql is required to run this script. Install the Postgres client (psql) and try again." >&2
  exit 1
fi

echo "Using DB URL: $DB_URL"
echo "Seeding test data..."

psql "$DB_URL" -v ON_ERROR_STOP=1 <<'SQL'
-- Create a test user
INSERT INTO users (name)
VALUES ('testuser')
ON CONFLICT (name) DO NOTHING;

-- Create some feeds owned by the test user (use the test user as owner)
INSERT INTO feeds (name, url, user_id)
VALUES
  ('TechCrunch', 'https://techcrunch.com/feed/', (SELECT id FROM users WHERE name = 'testuser'))
ON CONFLICT (url) DO NOTHING;

INSERT INTO feeds (name, url, user_id)
VALUES
  ('Hacker News', 'https://news.ycombinator.com/rss', (SELECT id FROM users WHERE name = 'testuser'))
ON CONFLICT (url) DO NOTHING;

INSERT INTO feeds (name, url, user_id)
VALUES
  ('boot.dev', 'https://blog.boot.dev/index.xml', (SELECT id FROM users WHERE name = 'testuser'))
ON CONFLICT (url) DO NOTHING;

-- Ensure the test user follows those feeds
INSERT INTO feed_follows (user_id, feed_id)
SELECT u.id, f.id
FROM users u
JOIN feeds f ON f.url IN (
  'https://techcrunch.com/feed/',
  'https://news.ycombinator.com/rss',
  'https://blog.boot.dev/index.xml'
)
WHERE u.name = 'testuser'
ON CONFLICT (user_id, feed_id) DO NOTHING;

-- Insert a couple of posts per feed (idempotent by url)
INSERT INTO posts (feed_id, title, url, published_at, description)
SELECT f.id,
       'Example post 1',
       CONCAT(f.url, 'example-post-1'),
       NOW() - INTERVAL '2 days',
       'Seeded example post'
FROM feeds f
WHERE f.url = 'https://techcrunch.com/feed/'
ON CONFLICT (url) DO NOTHING;

INSERT INTO posts (feed_id, title, url, published_at, description)
SELECT f.id,
       'Example post 2',
       CONCAT(f.url, 'example-post-2'),
       NOW() - INTERVAL '1 day',
       'Seeded example post 2'
FROM feeds f
WHERE f.url = 'https://news.ycombinator.com/rss'
ON CONFLICT (url) DO NOTHING;

INSERT INTO posts (feed_id, title, url, published_at, description)
SELECT f.id,
       'Example post 3',
       CONCAT(f.url, 'example-post-3'),
       NOW() - INTERVAL '3 days',
       'Seeded example post 3'
FROM feeds f
WHERE f.url = 'https://blog.boot.dev/index.xml'
ON CONFLICT (url) DO NOTHING;
SQL

echo "Seeding complete."

cat <<'INFO'

Next steps:
- Set the seeded user as the current user for the CLI:
    just run login testuser
  or:
    go run . login testuser
  or edit ~/.config/rbg/config.json setting "current_user_name": "testuser"

- Run the browse command to see seeded posts:
    just run browse
    # or
    go run . browse

To make this script callable from the Justfile, add a recipe that runs:
    ./scripts/seed_example.sh

INFO
