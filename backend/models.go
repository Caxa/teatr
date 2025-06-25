package backend

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

// Новые структуры для данных
type Performance struct {
	ID          int
	Title       string
	Description string
	Duration    int
	AgeRating   string
}

type Scene struct {
	ID       int
	Name     string
	Capacity int
	Address  string
}

type Actor struct {
	ID        int
	FullName  string
	BirthDate string
	Bio       string
}
