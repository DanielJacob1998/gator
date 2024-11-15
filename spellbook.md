---
title: Building CLI Commands in Go
tags:
  - command line interface
  - cli
  - go
  - golang
  - application state
---
## Summary

In Go, you can build a custom command-line interface (CLI) application by defining commands and handlers. Each command has a name and a set of arguments. The `os.Args` slice is used to access command-line arguments. A flexible command system involves creating a `command` struct to store command details and a `commands` struct to map command names to handler functions. Handlers modify application state, such as a configuration, based on command inputs. CLI applications can be built without external frameworks by manually parsing commands and implementing handler functions for each command.

## Docs

- [Go os Package: Args](https://pkg.go.dev/os#pkg-variables)
- [Go Programming Language: Command-Line Arguments](https://gobyexample.com/command-line-arguments)

## Examples

**Defining a basic command handler:**

```go
type state struct {
    config *config // Pointer to a configuration struct
}

type command struct {
    name string
    args []string
}

type commands struct {
    handlers map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) == 0 {
        return fmt.Errorf("a username is required")
    }
    username := cmd.args[0]
    s.config.User = username
    fmt.Printf("User set to %s\n", username)
    return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
    c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
    if handler, exists := c.handlers[cmd.name]; exists {
        return handler(s, cmd)
    }
    return fmt.Errorf("command not found")
}
```

**Using `os.Args` to parse commands:**

```go
func main() {
    if len(os.Args) < 2 {
        fmt.Println("Not enough arguments")
        os.Exit(1)
    }

    commandName := os.Args[1]
    args := os.Args[2:]

    stateInstance := &state{config: &config{}}
    commandInstance := command{name: commandName, args: args}

    commandHandlers := &commands{handlers: make(map[string]func(*state, command) error)}
    commandHandlers.register("login", handlerLogin)

    if err := commandHandlers.run(stateInstance, commandInstance); err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }
}
```

These examples illustrate how to set up a simple CLI application in Go, allowing commands to interact with application state using custom handler functions.
