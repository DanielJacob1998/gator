package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "context"
    
    _ "github.com/lib/pq"
    "github.com/DanielJacob1998/gator/internal/config"
    "github.com/DanielJacob1998/gator/internal/database"
)

type state struct {
    db  *database.Queries
    cfg *config.Config
}

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintf(os.Stderr, "Error: not enough arguments\n")
        os.Exit(1)
    }

    cfg, err := config.Read()
    if err != nil {
        log.Fatalf("error reading config: %v", err)
    }

    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("error opening db: %v", err)
    }
    dbQueries := database.New(db)

    programState := &state{
        cfg: &cfg,
        db:  dbQueries,
    }

    // Create a new context for your operations
    ctx := context.Background()

    cmds := commands{
        registeredCommands: make(map[string]func(*state, command) error),
    }
    cmds.register("login", handlerLogin)
    cmds.register("register", handlerRegister)
    cmds.register("users", usersHandler)
    cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
    cmds.register("agg", handleAgg)
    cmds.register("follow", middlewareLoggedIn(handleFollow))
    cmds.register("following", middlewareLoggedIn(followingCommand))
    cmds.register("feeds", feedsHandler)
    cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
    cmds.register("browse", middlewareLoggedIn(handlerBrowse))
    cmds.register("reset", handlerReset)

    cmdName := os.Args[1]
    cmdArgs := os.Args[2:]

    if err := cmds.run(programState, command{Name: cmdName, Args: cmdArgs}); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
