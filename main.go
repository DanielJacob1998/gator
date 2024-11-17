package main

import (
    "database/sql"
    "fmt"
    "log"
    "strings"
    "os"
    
    _ "github.com/lib/pq"
    "github.com/jackc/pgx/v4"
    "github.com/jackc/pgx/v4/stdlib"
    "github.com/DanielJacob1998/gator/internal/config"
    "github.com/DanielJacob1998/gator/internal/database"
)

type state struct {
    db  *database.Queries
    cfg *config.Config
}

func main() {
    // Check if we have enough arguments
    if len(os.Args) < 2 {
        fmt.Fprintf(os.Stderr, "Error: not enough arguments\n")
        os.Exit(1)
    }

    // Read config first
    cfg, err := config.Read()
    if err != nil {
        log.Fatalf("error reading config: %v", err)
    }

    // Add sslmode=disable to database URL
    dbURL := cfg.DatabaseURL
    /*
    if !strings.Contains(dbURL, "host=") {
        if strings.Contains(dbURL, "?") {
            dbURL += "&host=/tmp"
        } else {
            dbURL += "?host=/tmp"
        }
    }
    if !strings.Contains(dbURL, "sslmode=") {
        dbURL += "&sslmode=disable"
    }
    */
    // Open database connection
    connConfig, err := pgx.ParseConfig(dbURL)
    if err != nil {
        log.Fatal(err)
    }
    connConfig.TLSConfig = nil // Explicitly disable SSL
    db, err := sql.OpenDB(stdlib.OpenDB(*connConfig))
    if err != nil {
        log.Fatal(err)
    }
    
    dbQueries := database.New(db)

    programState := &state{
        cfg: &cfg,
        db:  dbQueries,
    }
    cmds := commands{
        registeredCommands: make(map[string]func(*state, command) error),
    }
    cmds.register("login", handlerLogin)
    cmds.register("register", handlerRegister)  // Move this line up!

    // Get command name and args
    cmdName := os.Args[1]
    cmdArgs := os.Args[2:]

    // Run the command
    if err := cmds.run(programState, command{Name: cmdName, Args: cmdArgs}); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
