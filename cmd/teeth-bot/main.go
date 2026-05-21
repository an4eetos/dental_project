package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/anuarkuanysh/dental_project/infra/app"
)

func main() {
	_ = godotenv.Load()
	fx.New(app.Module()).Run()
}
