package tests

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestMigrationsRollbackInTransaction(t *testing.T) {
	integrationEnabled(t)

	db := openPostgres(t)
	defer db.Close()

	schemas := []string{"identity", "logistics", "mobility", "payment", "operations"}
	for _, schema := range schemas {
		t.Run(schema, func(t *testing.T) {
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatalf("begin tx: %v", err)
			}
			defer tx.Rollback()

			files := migrationFiles(t, filepath.Join("..", "infra", "migrations", schema))
			for _, file := range files {
				content, err := os.ReadFile(file)
				if err != nil {
					t.Fatalf("read migration %s: %v", file, err)
				}
				if _, err := tx.ExecContext(context.Background(), string(content)); err != nil {
					t.Fatalf("exec migration %s: %v", filepath.Base(file), err)
				}
			}

			ctx, cancel := testContext(t, 5*time.Second)
			defer cancel()
			var marker int
			if err := tx.QueryRowContext(ctx, "SELECT 1").Scan(&marker); err != nil {
				t.Fatalf("migration transaction query: %v", err)
			}
			if marker != 1 {
				t.Fatalf("expected query marker 1, got %d", marker)
			}
		})
	}
}

func migrationFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(files)
	return files
}
