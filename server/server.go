package server

import "github.com/lng50k/booster-backend/config"

func Init() {
	config := config.GetConfig()
	r := NewRouter()
	r.Run(config.GetString("server.port"))
}
