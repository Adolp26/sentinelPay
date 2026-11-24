package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// 3. COMO USAR NO main.go
// -----------------------------------------------------------------------------
//   - Chame database.NewPool logo depois do config.Load()
//   - Erro → logger.Error(...) + os.Exit(1)  (sem banco não adianta subir)
//   - defer pool.Close()  → fecha as conexões quando o main terminar
//
//   💡 Macete: use um context com timeout (ex.: 5s) nessa chamada.
//   Se o banco estiver travado, a API falha rápido em vez de ficar pendurada.
//
// -----------------------------------------------------------------------------
// 4. NÃO FAÇA (ainda)
// -----------------------------------------------------------------------------
//   - Não guarde o pool numa variável global. Passe ele como parâmetro
//     para quem precisa (handlers, repositórios). Isso facilita testes.
//   - Não configure MaxConns, timeouts do pool etc. agora. Versão simples
//     primeiro; refinamos quando houver motivo.
// =============================================================================
