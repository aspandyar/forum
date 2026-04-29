package main

import "github.com/aspandyar/forum/internal/config/envfile"

func LoadEnvFromFile(filename string) error {
	return envfile.Load(filename)
}
