package main

func main() {
	a := App{}
	a.Initialize()

	a.Run(":" + a.conf.Port)
}
