package backend

import (
	"database/sql"
	"time"
)

type Seat struct {
	Seat   string
	Price  int
	IsFree bool
}

// Структура для хранения данных о билете
type Ticket struct {
	ID        int
	Price     int
	Seat      int
	OwnerName sql.NullString
	PosterID  int
}
type PageData struct {
	PerformanceTitle string
	StartTime        string
	SceneName        string
	MinPrice         int
	MaxPrice         int
	PosterID         string
	Seats            []Seat
}

type Play struct {
	ID     int
	Author string
	Title  string
}
type BookingConfirmation struct {
	OrderID          string
	PerformanceTitle string
	StartTime        time.Time
	SceneName        string
	Seats            []int
	TotalPrice       int
	CustomerName     string
	CustomerEmail    string
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

type ActorRole struct {
	ID                int
	ActorID           int
	PosterID          int
	PerformanceRoleID int
}

// Performance структура для хранения данных о спектакле
type Performance struct {
	ID          int
	Title       string
	Description string
	Duration    int
	AgeRating   string
}

// Actor структура для хранения данных об актере
type Actor struct {
	ID       int
	FullName string
	Troupe   string
}
type Scene struct {
	ID       int
	Name     string
	Capacity int
	Address  string
}
