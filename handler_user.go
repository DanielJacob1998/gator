package main

import (
    "context"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
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

    // Check if user exists first
    _, err := s.db.GetUser(context.Background(), name)
    if err == nil {
        return fmt.Errorf("user %s already exists", name)
    }

    // Create a new user in the database
    user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
        ID:        uuid.New(),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        Name:      name,
    })
    if err != nil {
        return fmt.Errorf("couldn't create user: %w", err)
    }

    if err := s.cfg.SetUser(name); err != nil {
        return fmt.Errorf("couldn't set current user: %w", err)
    }

    fmt.Printf("User created successfully: %+v\n", user)
    return nil
}
