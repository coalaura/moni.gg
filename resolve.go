package main

type Task interface {
	Resolve() StatusEntry
}
