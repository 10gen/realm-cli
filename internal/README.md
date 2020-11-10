# Rough Idea of Codebase

Will remove this file eventually, just putting it here for informational purposes early on.

```cmd
.
├── cmd                    # CLI command definitions
│   ├── app_create.go
│   ├── app_init.go
│   ├── ...
│   └── whoami.go
├── internal
│   ├── cli                # basic CLI infrastructure
│   ├── cmd
│   │   ├── cmd.go         # basic command infrastructure
│   │   ├── app
│   │   │   ├── create.go  # any 'app create' command-specific logic
│   │   │   ├── ...
│   │   │   └── list.go
│   │   ├── ...
│   │   └── whoami
│   │   │   └── whoami.go
│   ├── mock               # mocks of anything (here for access to internal/)
│   ├── utils              # any non-CLI related utilities
└── main.go
```
