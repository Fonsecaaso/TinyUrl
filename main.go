package main

func main() {
	redisClient := newRedisClient()
	r := setupRouter(redisClient)
	r.Run(":8080")
}
