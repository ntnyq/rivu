package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/updater"
)

func main() {
	store, err := config.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	timeout, err := time.ParseDuration(os.Args[2])
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	service := updater.New(store, nil)
	err = service.Check(ctx)
	_ = json.NewEncoder(os.Stdout).Encode(service.Snapshot())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
