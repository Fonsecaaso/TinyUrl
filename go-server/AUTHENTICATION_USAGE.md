# Guia de Uso - Autenticação JWT

## Resumo das Melhorias

### ✅ Problemas Resolvidos

1. **Type-Safety**: Agora usamos `CustomClaims` com tipos definidos ao invés de `map[string]interface{}`
2. **Middleware Gin**: Migrado de `http.Handler` para `gin.HandlerFunc`
3. **Claims no Contexto**: Middleware armazena automaticamente as claims e user_id no contexto
4. **Helper Functions**: Funções utilitárias para extrair dados do contexto de forma segura
5. **Consistência**: Token agora usa `user_id` tanto na geração quanto na validação

### 📁 Arquivos Modificados

- `internal/token/jwt.go` - Estrutura CustomClaims e funções de token
- `internal/middleware/auth_middleware.go` - Middleware Gin e helpers
- `internal/handler/url_handler.go` - Handler simplificado
- `internal/routes/routes.go` - Rotas protegidas

## Como Usar

### 1. No Handler (Forma Recomendada)

```go
func (h *URLHandler) GetUserURLs(c *gin.Context) {
    // Extrai user_id de forma type-safe
    userID, err := middleware.GetUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // userID já é do tipo uuid.UUID, pronto para usar!
    urls, err := h.service.GetUserURLs(c.Request.Context(), userID)
    // ...
}
```

### 2. Acessar Claims Completas (Opcional)

Se precisar de outras informações das claims:

```go
func (h *Handler) SomeFunction(c *gin.Context) {
    claims, err := middleware.GetClaimsFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Acessa os campos
    userID := claims.UserID
    expiresAt := claims.ExpiresAt
    issuedAt := claims.IssuedAt
    // ...
}
```

### 3. Proteger Rotas

No arquivo `routes.go`:

```go
// Rotas protegidas
protected := api.Group("/user")
protected.Use(middleware.AuthMiddleware())
{
    protected.GET("/urls", urlHandler.GetUserURLs)
    protected.POST("/profile", userHandler.UpdateProfile)
    // ... outras rotas protegidas
}
```

### 4. Exemplo de Requisição

```bash
# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# Response: { "token": "eyJhbGc..." }

# Acessar rota protegida
curl -X GET http://localhost:8080/api/user/urls \
  -H "Authorization: Bearer eyJhbGc..."
```

## Vantagens da Nova Abordagem

### Antes (❌)
```go
// Muitas verificações manuais e type assertions
claims, exists := c.Get("claims")
if !exists { /* ... */ }

userClaims, ok := claims.(map[string]interface{})
if !ok { /* ... */ }

userIDStr, ok := userClaims["user_id"].(string)
if !ok || userIDStr == "" { /* ... */ }

userID, err := uuid.Parse(userIDStr)
if err != nil { /* ... */ }
```

### Agora (✅)
```go
// Uma linha, type-safe, com error handling
userID, err := middleware.GetUserIDFromContext(c)
if err != nil { /* ... */ }
```

## Benefícios

1. **Código Mais Limpo**: Menos boilerplate em cada handler
2. **Type-Safety**: Compilador verifica os tipos em tempo de compilação
3. **Facilita Manutenção**: Mudanças na estrutura de claims em um único lugar
4. **Melhor Testabilidade**: Helpers podem ser facilmente mockados
5. **Reutilizável**: Mesma lógica em todos os handlers protegidos
6. **Consistente**: Token sempre usa os mesmos campos (user_id)

## Estrutura do Token JWT

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "exp": 1234567890,
  "iat": 1234567890,
  "nbf": 1234567890
}
```

## Próximos Passos Recomendados

1. ✅ **Mover secret para variável de ambiente**
   - Atualmente hardcoded em `token/jwt.go`

2. ✅ **Adicionar refresh tokens**
   - Implementar tokens de longa duração

3. ✅ **Rate limiting por usuário**
   - Usar user_id do contexto para limitar requisições

4. ✅ **Logging melhorado**
   - Incluir user_id em todos os logs de rotas protegidas

5. ✅ **Testes unitários**
   - Testar middleware e helpers com casos de erro
