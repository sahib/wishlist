package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func withTestDatabase(t *testing.T, fn func(db *Database)) {
	path := "/tmp/anon-test.db"
	defer os.Remove(path)

	db, err := NewDatabase(path)
	require.Nil(t, err)

	fn(db)
}

func TestDatabaseOpen(t *testing.T) {
	withTestDatabase(t, func(db *Database) {
		fmt.Println(db)
	})
}

func TestDatabaseGetItems(t *testing.T) {
	withTestDatabase(t, func(db *Database) {
		items, err := db.GetItems()
		require.Nil(t, err)
		require.Equal(t, []*Item{}, items)

		for idx := 0; idx < 10; idx++ {
			itemID, err := db.AddItem(fmt.Sprintf("%d", idx), "http://something")
			require.Nil(t, err)
			require.Equal(t, int64(idx+1), itemID)
		}

		items, err = db.GetItems()
		require.Nil(t, err)
		require.Len(t, items, 10)

		for idx, item := range items {
			require.Equal(t, int64(idx+1), item.ID)
			require.Equal(t, fmt.Sprintf("%d", idx), item.Name)
			require.Equal(t, "http://something", item.Link)
		}
	})
}
