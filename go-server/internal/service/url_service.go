package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Body struct {
	URL string `json:"url"`
}

func CreateTinyUrl(c *gin.Context, redisClient *redis.Client, pgClient *pgxpool.Pool) {
	ctx := context.Background()
	body := Body{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var key string
	var err error
	for {
		key = createId()
		var count int
		err = pgClient.QueryRow(ctx, "SELECT COUNT(*) FROM urls WHERE id = $1", key).Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar ID no banco"})
			return
		}
		if count == 0 {
			break
		}
	}

	_, err = pgClient.Exec(ctx, "INSERT INTO urls (id, url) VALUES ($1, $2)", key, body.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao armazenar a URL no banco"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL armazenada com sucesso", "chave gerada": key})
}

func GetUrl(c *gin.Context, redisClient *redis.Client, pgClient *pgxpool.Pool) {
	ctx := context.Background()
	id := c.Param("id")

	val, err := redisClient.Get(ctx, id).Result()
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Valor encontrado no Redis", "id": id, "value": val})
		return
	}

	if err != redis.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar chave no Redis"})
		return
	}

	var url string
	err = pgClient.QueryRow(ctx, "SELECT url FROM urls WHERE id = $1", id).Scan(&url)
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "ID não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao consultar o banco de dados"})
		return
	}

	err = redisClient.Set(ctx, id, url, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar no Redis"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Valor encontrado no PostgreSQL", "id": id, "value": url})
}

func createId() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, 6)

	for i := 0; i < 6; i++ {
		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		id[i] = chars[randomIndex.Int64()]
	}

	return string(id)
}
