# BlogGator - RSS Feed Aggregator CLI

BlogGator is a command-line RSS feed aggregator that allows you to follow and browse RSS feeds directly from your terminal. Built with Go and PostgreSQL.

## Prerequisities
Before running BlogGator, you'll need to have the following installed:
- **Go** (version 1.21 or higher) - [Download here](https://golang.org/dl/)
- **PostgreSQL** - [Download here](https://www.postgresql.org/download/)

Install the Gator CLI using Go's install command:
```bash
go install github.com/yourusername/gator@latest
```

# Setup
## 1. Create a PostgresSQL Database 
Create a new database for BlogGator
```bash
    createdb gator
```


## 2. Configure Gator
Create a .gatorconfig.json file in your home directory with the following structure:

``` bash
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace username and password with your PostgreSQL credentials.

## 3. Run Migrations
The database schema will be automatically created when you first run the program.

# Usage
## Register a User
``` bash
gator register <username>
```

This creates a new user and automatically logs you in.

## Add RSS Feeds
``` bash
gator addfeed <feed_name> <feed_url>
```

Example:

gator addfeed "Tech cruch blog Blog" https://techcrunch.com/feed/

## Follow a Feed
``` bash
gator follow <feed_url>
```

## View Your Followed Feeds
``` bash
gator following
```

## Aggregate Posts
Start fetching posts from your followed feeds:
``` bash
gator agg <time_between_requests>
```
Example (fetch every 1 minute):

``` bash
gator agg 1m
```

## Browse Posts
View recent posts from your followed feeds:

```bash
gator browse [limit]
```

Example (show 10 most recent posts):

``` bash
gator browse 10
```


## Other Commands

``` gator login <username> ``` - Switch to a different user
``` gator users ``` - List all registered users
``` gator feeds ``` - List all available feeds
``` gator unfollow <feed_url> ``` - Unfollow a feed
``` gator reset ``` - Reset the database (warning: deletes all data!)


# Example Workflow
``` bash
## Register yourself
gator register alice

# Add some feeds
gator addfeed "Hacker News" https://hnrss.org/frontpage
gator addfeed "Go Blog" https://go.dev/blog/feed.atom

# Follow the feeds
gator follow https://hnrss.org/frontpage
gator follow https://go.dev/blog/feed.atom

# Start aggregating (in a separate terminal)
gator agg 5m

# Browse your posts
gator browse 5

```