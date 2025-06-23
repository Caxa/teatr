package main

import (
	"time"
)

type Play struct {
	ID     int
	Author string
	Title  string
}

type Director struct {
	ID       int
	FullName string
}

type Scene struct {
	ID       int
	Name     string
	Capacity int
	Address  string
}

type Performance struct {
	ID     int
	Title  string
	PlayID int
}

type Actor struct {
	ID       int
	FullName string
	Troupe   string
}

type PerformanceRole struct {
	ID            int
	Name          string
	PerformanceID int
	ActorID       int
}

type Poster struct {
	ID            int
	StartTime     time.Time
	SceneID       int
	DirectorID    int
	PerformanceID int
}

type Ticket struct {
	ID        int
	RoleName  string
	Price     int
	Seat      int
	OwnerName string
	PosterID  int
}

type ActorRole struct {
	ID                int
	ActorID           int
	PosterID          int
	PerformanceRoleID int
}
