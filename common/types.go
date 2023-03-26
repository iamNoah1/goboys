package common

type Cowboy struct {
	Name   string `json:"name"`
	Health int    `json:"health"`
	Damage int    `json:"damage"`
}

type ShootingData struct {
	Source string `json:"source"`
	Damage int    `json:"damage"`
}

type HealthData struct {
	Name   string `json:"name"`
	Health int    `json:"health"`
}
