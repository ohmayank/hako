# hako

Drop a file into a remote mount. If the remote is down, it queues locally and retries until it goes through.

## Requirements

- Go 1.25 or newer

## Build

```sh
make build
```

This creates a local `bin/hako` executable.

## Usage

Ingest a file with default paths:

```sh
bin/hako ingest ./report.pdf
```

Use custom remote and queue paths:

```sh
bin/hako ingest ./report.pdf --remote /mnt/nfs/uploads --queue ./queue
```

Run the retry worker:

```sh
bin/hako worker --remote /mnt/nfs/uploads --queue ./queue
```

Use a custom retry interval:

```sh
bin/hako worker --remote /mnt/nfs/uploads --queue ./queue --interval-seconds 10
```

Run through Make:

```sh
make run ARGS='ingest ./report.pdf --remote ./remote --queue ./queue'
```

## How It Works

```text
source file
    |
    v
atomic copy to remote
    |
    +-- success --> remote/file
    |
    +-- failure --> queue/timestamp_random_file
                         |
                         v
                    worker retry loop
                         |
                         v
                    remote/file
```

## Development

Run the standard checks:

```sh
make check
```

Individual commands:

```sh
make test
make race
make vet
make fmt-check
```
