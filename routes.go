package main

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func setupRouter(redis *redis.Client) *gin.Engine {
	r := gin.Default()

	r.POST("/:url", func(c *gin.Context) {
		url := c.Param("url")
		key := hashURL(url)
		err := redis.Set(ctx, key, url, 0).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao armazenar a URL no Redis"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "URL armazenada com sucesso", "chave gerada": key})
	})

	r.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		val, err := redis.Get(ctx, id).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar chave no Redis"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Valor encontrado", "id": id, "value": val})
	})

	return r
}

func hashURL(url string) string {
	// Cria o hash SHA-256 da URL
	hash := sha256.New()
	hash.Write([]byte(url))

	// Converte o hash para base64
	hashBytes := hash.Sum(nil)
	base64Hash := base64.URLEncoding.EncodeToString(hashBytes)

	// Retorna os primeiros 6 caracteres do hash base64
	// A string gerada estará em base64 e usaremos a versão URL-safe
	shortenedHash := base64Hash[:6]

	return shortenedHash
}
