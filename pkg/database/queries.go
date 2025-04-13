package database

import (
	"context"
	"fmt"
	"strings"
)

// Função para atualizar um registro em uma tabela específica no banco de dados
func Update(ctx context.Context, db DB, tableName, id string, fields map[string]any) error {
	if len(fields) == 0 {
		return fmt.Errorf("nenhum campo fornecido para atualização")
	}

	setters := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)
	args = append(args, id)

	i := 2
	for field, value := range fields {
		setters = append(setters, fmt.Sprintf("%s = $%d", field, i))
		args = append(args, value)
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $1", tableName, strings.Join(setters, ", "))
	_, err := db.ExecContext(ctx, query, args...)
	return err
}
