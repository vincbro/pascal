package blaise

import "fmt"

type Time uint32

func (t Time) ToHMSString() string {
	h := t / 3600
	m := (t % 3600) / 60
	s := t % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

type Location struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Coordinate Coordinate `json:"coordinate"`
}

type Coordinate struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type Itenirary struct {
	From          Location `json:"from"`
	To            Location `json:"to"`
	DepartureTime Time     `json:"departue_time"`
	ArrivalTime   Time     `json:"arrival_time"`
	Legs          []Leg    `json:"legs"`
}

type Leg struct {
	From          Location `json:"from"`
	To            Location `json:"to"`
	DepartureTime Time     `json:"departue_time"`
	ArrivalTime   Time     `json:"arrival_time"`
	Stops         []Stop   `json:"stops"`
	Shapes        []Shape  `json:"shapes"`
	Mode          string   `json:"mode"`
	HeadSign      *string  `json:"head_sign"`
	LongName      *string  `json:"long_name"`
	ShortName     *string  `json:"short_name"`
}

type Stop struct {
	Location        Location `json:"location"`
	DepartureTime   Time     `json:"departue_time"`
	ArrivalTime     Time     `json:"arrival_time"`
	DistanceTraveld float32  `json:"distance_traveld"`
}

type Shape struct {
	Location        Location `json:"location"`
	Sequence        uint     `json:"sequence"`
	DistanceTraveld float32  `json:"distance_traveld"`
}
