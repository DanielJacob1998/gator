package main

import (
    "context"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
    "github.com/lib/pq"
    "os"
    "github.com/google/uuid"
    "github.com/DanielJacob1998/gator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
    if len(cmd.Args) != 1 {
        return fmt.Errorf("usage: %s <name>", cmd.Name)
    }
    name := cmd.Args[0]

    // Check if user exists in database
    _, err := s.db.GetUser(context.Background(), name)
    if err != nil {
        return fmt.Errorf("user does not exist")
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

func handlerReset(ctx context.Context, s state) {
    if err := resetTable(ctx, s.db); err != nil {
        fmt.Println("Failed to reset database:", err)
        os.Exit(1) // Non-zero exit code for failure
    }
    fmt.Println("Database reset successful.")
    os.Exit(0) // Zero exit code for success
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
