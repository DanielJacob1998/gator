package main

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    "os"
    "log"
    "strings"
    "strconv"
    
    _ "github.com/lib/pq"
    "github.com/lib/pq"
    "github.com/google/uuid"
    "github.com/DanielJacob1998/gator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
    if len(cmd.Args) != 1 {
        return fmt.Errorf("usage: %v <name>", cmd.Name)
    }
    name := cmd.Args[0]

    _, err := s.db.GetUser(context.Background(), name)
    if err != nil {
        return fmt.Errorf("couldn't find user: %w", err)
    }

    err = s.cfg.SetUser(name)
    if err != nil {
        return fmt.Errorf("couldn't set current user: %w", err)
    }

    fmt.Println("User switched successfully!")
    return nil
}

func handlerRegister(s *state, cmd command) error {
    if len(cmd.Args) != 1 {
        return fmt.Errorf("usage: %s <name>", cmd.Name)
    }
    name := cmd.Args[0]

    // Try to create user directly
    user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
        ID:        uuid.New(),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        Name:      name,
    })
    
    // Handle potential duplicate user error
    if err != nil {
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
            fmt.Fprintf(os.Stderr, "user %s already exists\n", name)
            os.Exit(1) // This is what makes the program exit with status 1
        }
        return fmt.Errorf("couldn't create user: %w", err)
    }

    if err := s.cfg.SetUser(name); err != nil {
        return fmt.Errorf("couldn't set current user: %w", err)
    }

    fmt.Printf("User created successfully: %+v\n", user)
    return nil
}

func resetTable(ctx context.Context, db *database.Queries) error {
    return db.DeleteUsers(ctx)
}

func handlerReset(s *state, cmd command) error {
    err := s.db.DeleteUsers(context.Background())
    if err != nil {
        return fmt.Errorf("couldn't delete users: %w", err)
    }
    fmt.Println("Database reset successfully!")
    return nil
}

func usersHandler(s *state, c command) error {
    // Get all users
    users, err := s.db.GetUsers(context.Background())
    if err != nil {
        return err
    }

    // Get current user from config
    currentUser := s.cfg.CurrentUserName

    // Print each user
    for _, username := range users {
        if username == currentUser {
            fmt.Printf("* %s (current)\n", username)
        } else {
            fmt.Printf("* %s\n", username)
        }
    }

    return nil
}

func handleAgg(s *state, cmd command) error {
    if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
        return fmt.Errorf("usage: %v <time_between_reqs>", cmd.Name)
    }

    timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
    if err != nil {
        return fmt.Errorf("invalid duration: %w", err)
    }

    log.Printf("Collecting feeds every %s...", timeBetweenRequests)

    ticker := time.NewTicker(timeBetweenRequests)

    for ; ; <-ticker.C {
        scrapeFeeds(s)
    }
}

func addfeed(db *sql.DB, name string, url string, userID uuid.UUID) error {
    ctx := context.Background()
    q := database.New(db)

    // First add the feed
    params := database.AddFeedParams{
        ID:        uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Name:      name,
        Url:       url,
        UserID:    userID,
    }

    feed, err := q.AddFeed(ctx, params)
    if err != nil {
        return fmt.Errorf("couldn't add feed: %v", err)
    }

    // Then create a feed follow record
    followParams := database.CreateFeedFollowParams{
        ID:        uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        UserID:    userID,
        FeedID:    feed.ID,  // Use the ID of the feed we just created
    }

    _, err = q.CreateFeedFollow(ctx, followParams)
    if err != nil {
        return fmt.Errorf("couldn't create feed follow: %v", err)
    }

    fmt.Printf("Feed added: Name=%s, URL=%s\n", name, url)
    return nil
}

func getCurrentUserID(s *state) (uuid.UUID, error) {
    userName := s.cfg.CurrentUserName
    // Retrieve the user ID from the database using the userName
    user, err := s.db.GetUser(context.Background(), userName)
    if err != nil {
        return uuid.Nil, err
    }
    return user.ID, nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
    return func(s *state, cmd command) error {
        user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
        if err != nil {
            return fmt.Errorf("user not logged in: %v", err)
        }
        return handler(s, cmd, user)
    }
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
    if len(cmd.Args) != 2 {
        return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
    }

    name := cmd.Args[0]
    url := cmd.Args[1]

    feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
        ID:        uuid.New(),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        UserID:    user.ID,
        Name:      name,
        Url:       url,
    })
    if err != nil {
        return fmt.Errorf("couldn't create feed: %w", err)
    }

    feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
        ID:        uuid.New(),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        UserID:    user.ID,
        FeedID:    feed.ID,
    })
    if err != nil {
        return fmt.Errorf("couldn't create feed follow: %w", err)
    }

    fmt.Println("Feed created successfully:")
    printFeed(feed, user)
    fmt.Println()
    fmt.Println("Feed followed successfully:")
    printFeedFollow(feedFollow.UserName, feedFollow.FeedName)
    fmt.Println("=====================================")
    return nil
}

func feedsHandler(s *state, c command) error {
    feeds, err := s.db.GetAllFeeds(context.Background())
    if err != nil {
        return err
    }

    for _, feed := range feeds {
        fmt.Printf("Feed: %s\nURL: %s\nCreated by: %s\n\n",
            feed.Name,
            feed.Url,
            feed.CreatorName)
    }
    return nil
}

func handleFollow(s *state, c command, user database.User) error {
    if len(c.Args) != 1 {
        return fmt.Errorf("follow command requires exactly 1 argument: URL")
    }
    url := c.Args[0]

    // Use the user parameter directly
    // Get feed by URL
    feed, err := s.db.GetFeedByURL(context.Background(), url)
    if err != nil {
        return fmt.Errorf("error getting feed: %w", err)
    }

    // Create feed follow
    params := database.CreateFeedFollowParams{
        ID:        uuid.New(),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        UserID:    user.ID,
        FeedID:    feed.ID,
    }
    
    feedFollow, err := s.db.CreateFeedFollow(context.Background(), params)
    if err != nil {
        return fmt.Errorf("error creating feed follow: %w", err)
    }

    fmt.Printf("Following feed '%v' for user '%v'\n", feedFollow.FeedName, feedFollow.UserName)
    return nil
}

func getAuthenticatedUser(s *state) (database.User, error) {
    if s.cfg.CurrentUserName == "" {
        return database.User{}, fmt.Errorf("not authenticated")
    }

    // Get user by name using the existing GetUser function
    user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
    if err != nil {
        return database.User{}, fmt.Errorf("not authenticated")
    }

    return user, nil
}

func followingCommand(s *state, c command, user database.User) error {
    // Use the user parameter directly

    feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
    if err != nil {
        return err
    }

    for _, follow := range feedFollows {
        fmt.Println(follow.FeedName)
    }

    return nil
}

func unfollowHandler(s *state, cmd command, user database.User) error {
    // Get the URL from the command Args
    if len(cmd.Args) < 1 {
        return fmt.Errorf("feed URL is required")
    }
    feedURL := cmd.Args[0]

    ctx := context.Background()
    err := s.db.UnfollowFeed(ctx, database.UnfollowFeedParams{
        UserID: user.ID,
        Url:    feedURL,
    })
    if err != nil {
        return fmt.Errorf("could not unfollow feed: %w", err)
    }

    return nil
}

func scrapeFeeds(s *state) {
    feed, err := s.db.GetNextFeedToFetch(context.Background())
    if err != nil {
        log.Println("Couldn't get next feeds to fetch", err)
        return
    }
    log.Println("Found a feed to fetch!")
    scrapeFeed(s.db, feed)
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
    _, err := db.MarkFeedFetched(context.Background(), feed.ID)
    if err != nil {
        log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
        return
    }

    feedData, err := fetchFeed(context.Background(), feed.Url)
    if err != nil {
        log.Printf("Couldn't collect feed %s: %v", feed.Name, err)
        return
    }
    for _, item := range feedData.Channel.Item {
        publishedAt := sql.NullTime{}
        if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
            publishedAt = sql.NullTime{
                Time:  t,
                Valid: true,
            }
        }

        _, err = db.CreatePost(context.Background(), database.CreatePostParams{
            ID:        uuid.New(),
            CreatedAt: time.Now().UTC(),
            UpdatedAt: time.Now().UTC(),
            FeedID:    feed.ID,
            Title:     item.Title,
            Description: sql.NullString{
                String: item.Description,
                Valid:  true,
            },
            Url:         item.Link,
            PublishedAt: publishedAt,
        })
        if err != nil {
            if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
                continue
            }
            log.Printf("Couldn't create post: %v", err)
            continue
        }
    }
    log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}

func handlerBrowse(s *state, cmd command, user database.User) error {
    limit := 2
    if len(cmd.Args) == 1 {
        if specifiedLimit, err := strconv.Atoi(cmd.Args[0]); err == nil {
            limit = specifiedLimit
        } else {
            return fmt.Errorf("invalid limit: %w", err)
        }
    }

    posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
        UserID: user.ID,
        Limit:  int32(limit),
    })
    if err != nil {
        return fmt.Errorf("couldn't get posts for user: %w", err)
    }

    fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
    for _, post := range posts {
        fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
        fmt.Printf("--- %s ---\n", post.Title)
        fmt.Printf("    %v\n", post.Description.String)
        fmt.Printf("Link: %s\n", post.Url)
        fmt.Println("=====================================")
    }

    return nil
}

func handlerListFeeds(s *state, cmd command) error {
    feeds, err := s.db.GetFeeds(context.Background())
    if err != nil {
        return fmt.Errorf("couldn't get feeds: %w", err)
    }

    if len(feeds) == 0 {
        fmt.Println("No feeds found.")
        return nil
    }

    fmt.Printf("Found %d feeds:\n", len(feeds))
    for _, feed := range feeds {
        user, err := s.db.GetUserById(context.Background(), feed.UserID)
        if err != nil {
            return fmt.Errorf("couldn't get user: %w", err)
        }
        printFeed(feed, user)
        fmt.Println("=====================================")
    }

    return nil
}

func printFeed(feed database.Feed, user database.User) {
    fmt.Printf("* ID:            %s\n", feed.ID)
    fmt.Printf("* Created:       %v\n", feed.CreatedAt)
    fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
    fmt.Printf("* Name:          %s\n", feed.Name)
    fmt.Printf("* URL:           %s\n", feed.Url)
    fmt.Printf("* User:          %s\n", user.Name)
    fmt.Printf("* LastFetchedAt: %v\n", feed.LastFetchedAt.Time)
}
