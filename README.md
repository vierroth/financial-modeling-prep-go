This library provides a thin go wrapper for the [FMP API](https://site.financialmodelingprep.com/developer/docs), providing a typesafe and easy to use interface.

## Usage

### Installation

The package is available on [GitHub](https://github.com/vierroth/financial-modeling-prep-go) and can be installed as follows:

```bash
go get github.com/vierroth/financial-modeling-prep-go@latest
```

### Setup

To begin with, create a client instance, providing the `accountId`, `API key` and transport:

```go
Client := fmp.New( "ApiKey", http.DefaultClient)
```

### Usage

```go
resp, err := Client.TrackSignUp(ctx, firstpromoter.TrackSignUpInput{
	Email: firstpromoter.String("test@test.com"),
	Tid:   firstpromoter.String("tid"),
})
```
