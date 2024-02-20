# lands

Package lands provides a small but nuanced helper function to support starting
and stopping http servers. Notably:

- It supports systemd socket activation.
- It supports `PORT` provided via the environment.
- It supports a default address configured in code.
- It supports graceful termination via Context.
- It supports graceful termination via `SIGINT` or `SIGTERM`.
