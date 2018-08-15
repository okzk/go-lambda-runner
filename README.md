# go-lambda-runner

A simple function runner for lambda function written in golang.

## Install 

```bash
go get -u github.com/okzk/go-lambda-runner
```

## Usage

Execute your function via go-lambda-runner.

```
go-lambda-runner go run main.go
```

If you need a input payload, use pipe.
```
cat input.json | go-lambda-runner go run main.go
```


## License

MIT
