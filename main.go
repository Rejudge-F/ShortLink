package main

func main() {
	app := App{}
	app.Initialize(getEnv())
	app.Run(":8000")
}
