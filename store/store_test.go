package store

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var DB_PATH = "/tmp/myTestDB.bolt"

func TestPut(t *testing.T) {
	s, err := initStore()
	defer os.Remove(DB_PATH)
	defer s.DB.Close()
	require.Nil(t, err)

	file := strings.NewReader("Next Meme")
	err = s.Put("public", "Ok, that was epic!", file)
	require.Nil(t, err)

	file = strings.NewReader("I have the highground!")
	err = s.Put("public", "///prequel////It's over Anakin!/////", file)
	require.Nil(t, err)
	file = strings.NewReader("Ackbar")
	err = s.Put("public", "/prequel/it's A trap", file)
	require.Nil(t, err)

	file = strings.NewReader("I have the highground!")
	err = s.Put("public", "It's over Anakin!", file)
	require.Nil(t, err)

	result, err := s.View("public")
	require.Nil(t, err)

	expected := "1\n" +
		"  2\n" +
		"It's over Anakin!\n" +
		"Ok, that was epic!\n" +
		"The Ring\n" +
		"a\n" +
		"  b\n" +
		"    c\n" +
		"      Hello there\n" +
		"prequel\n" +
		"  It's over Anakin!\n" +
		"  it's A trap\n"

	assert.Equal(t, expected, string(result))
}

func TestGet(t *testing.T) {
	tt := []struct {
		Collection string
		Keys       []string
		Result     string
	}{
		{
			"public",
			[]string{},
			"1\n  2\nThe Ring\na\n  b\n    c\n      Hello there\n",
		},
		{
			"public",
			[]string{"1"},
			"2\n",
		},
		{
			"public",
			[]string{"The Ring"},
			"My precious",
		},
		{
			"public",
			[]string{"a", "b", "c"},
			"Hello there\n",
		},
		{
			"public",
			[]string{"a", "b", "c", "Hello there"},
			"General Kenobi",
		},
		{
			"public",
			[]string{"invalid name kfj;lkdfj:"},
			"General Kenobi",
		},
	}

	s, err := initStore()
	require.Nil(t, err)
	defer s.DB.Close()
	defer s.Drop()
	require.Nil(t, err)

	for _, test := range tt {
		result, err := s.Get(test.Collection, test.Keys)
		require.Nil(t, err)

		assert.Equal(t, test.Result, string(result))
	}
	// 	s, err := initStore()
	// 	defer os.Remove(DB_PATH)
	// 	require.Nil(t, err)
	//
	// 	b, err := s.Get("public", "1/2")
	// 	require.Nil(t, err)
	// 	assert.Equal(t, "0", string(b))
	//
	// 	b, err = s.Get("public", "empty")
	// 	require.Nil(t, err)
	// 	assert.Equal(t, "", string(b))
}

func TestDelete(t *testing.T) {
	s, err := initStore()
	defer os.Remove(DB_PATH)
	defer s.DB.Close()
	require.Nil(t, err)

	err = s.Delete("public", "1/2")
	require.Nil(t, err)
	view, err := s.View("public")
	require.Nil(t, err)
	assert.Equal(t, "1\nThe Ring\na\n  b\n    c\n      Hello there\n", string(view))

	err = s.Delete("public", "a")
	require.Nil(t, err)
	view, err = s.View("public")
	require.Nil(t, err)
	assert.Equal(t, "1\nThe Ring\n", string(view))
}

// view of db
// 1
//   2
// The Ring
// a
//   b
//     c
// Hello there
func initStore() (*Store, error) {
	db, err := bolt.Open(DB_PATH, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	s := &Store{
		Path: DB_PATH,
		DB:   db,
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("public"))
		if err != nil {
			return err
		}

		b, err = b.CreateBucketIfNotExists([]byte("1"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("2"), []byte("0"))
		if err != nil {
			return err
		}

		b, err = tx.CreateBucketIfNotExists([]byte("public"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("The Ring"), []byte("My precious"))
		if err != nil {
			return err
		}

		// nested folders
		for _, v := range []string{"a", "b", "c"} {
			b, err = b.CreateBucketIfNotExists([]byte(v))
			if err != nil {
				return err
			}
		}

		err = b.Put([]byte("Hello there"), []byte("General Kenobi"))
		if err != nil {
			return err
		}
		return nil
	})
	return s, err
}
