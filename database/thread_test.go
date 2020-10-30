package database

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/fiatjaf/ilno/ilno"
)

func TestDatabase_Thread(t *testing.T) {
	t.Run("not exist", func(t *testing.T) {
		got, err := db.GetThreadByURI(context.Background(), "/not-exist")
		if !errors.Is(err, ilno.ErrStorageNotFound) {
			t.Errorf("Database.GetThreadByURI() error = %v, wantErr %v", err, ilno.ErrStorageNotFound)
			return
		}
		emptyt := ilno.Thread{}
		if !reflect.DeepEqual(got, emptyt) {
			t.Errorf("Database.GetThreadByURI() = %v, want %v", got, emptyt)
		}

		got, err = db.GetThreadByID(context.Background(), 1024)
		if !errors.Is(err, ilno.ErrStorageNotFound) {
			t.Errorf("Database.GetThreadByID() error = %v", err)
			return
		}
		if !reflect.DeepEqual(got, emptyt) {
			t.Errorf("Database.GetThreadByID() = %v, want %v", got, emptyt)
		}
	})
	t.Run("normal", func(t *testing.T) {
		newt1, err := db.NewThread(context.Background(), "/uri", "hello")
		if (err != nil) != false {
			t.Errorf("Database.NewThread() error = %v, wantErr %v", err, false)
			return
		}

		newt2, err := db.NewThread(context.Background(), "/about", "test")
		if (err != nil) != false {
			t.Errorf("Database.NewThread() %v", err)
			return
		}

		_, err = db.NewThread(context.Background(), "/about", "")
		if (err != nil) != true {
			t.Errorf("Database.NewThread() need error %v, but got nil", ilno.ErrInvalidParam)
			return
		}

		got, err := db.GetThreadByURI(context.Background(), "/uri")
		if err != nil {
			t.Errorf("Database.GetThreadByURI() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, newt1) {
			t.Errorf("Database.GetThreadByID() = %v, want %v", got, newt1)
		}

		got, err = db.GetThreadByID(context.Background(), newt2.ID)
		if err != nil {
			t.Errorf("Database.GetThreadByID() error = %v", err)
			return
		}
		if !reflect.DeepEqual(got, newt2) {
			t.Errorf("Database.GetThreadByID() = %v, want %v", got, newt2)
		}
	})
}
