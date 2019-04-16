package main

import "github.com/moonlightming/simple-single-log-center/controller"

func main()  {
	controller.InitWebServer()
	controller.InitRouter()
	controller.StartWebServer()
}
