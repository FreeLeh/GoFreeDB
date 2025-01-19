package freedb

import (
	"context"
	"fmt"

	"github.com/FreeLeh/GoFreeDB/google/auth"
)

func ExampleGoogleSheetRowStore() {
	// Initialize authentication
	googleAuth, err := auth.NewServiceFromFile(
		"<path_to_service_account_file>",
		GoogleAuthScopes,
		auth.ServiceConfig{},
	)
	if err != nil {
		panic(err)
	}

	// Create row store with columns definition
	store := NewGoogleSheetRowStore(
		googleAuth,
		"<spreadsheet_id>",
		"<sheet_name>",
		GoogleSheetRowStoreConfig{
			Columns: []string{"name", "age", "email"},
		},
	)

	// Insert some rows
	type Person struct {
		Name  string `db:"name"`
		Age   int    `db:"age"`
		Email string `db:"email"`
	}

	err = store.Insert(
		Person{Name: "Alice", Age: 30, Email: "alice@example.com"},
		Person{Name: "Bob", Age: 25, Email: "bob@example.com"},
	).Exec(context.Background())
	if err != nil {
		panic(err)
	}

	// Query rows
	var people []Person
	err = store.Select(&people).
		Where("age > ?", 20).
		OrderBy([]ColumnOrderBy{{Column: "age", OrderBy: OrderByAsc}}).
		Limit(10).
		Exec(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Selected people:", people)

	// Update rows
	update := map[string]interface{}{"age": 31}
	err = store.Update(update).Where("name = ?", "Alice").
		Exec(context.Background())
	if err != nil {
		panic(err)
	}

	// Count rows
	count, err := store.Count().
		Where("age > ?", 20).
		Exec(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Number of people over 20:", count)

	// Delete rows
	err = store.Delete().
		Where("name = ?", "Bob").
		Exec(context.Background())
	if err != nil {
		panic(err)
	}

	// Clean up
	err = store.Close(context.Background())
	if err != nil {
		panic(err)
	}
}
