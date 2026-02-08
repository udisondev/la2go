package model

import (
	"sync"
	"testing"
)

func TestNewWorldObject(t *testing.T) {
	loc := NewLocation(100, 200, 300, 1000)
	obj := NewWorldObject(12345, "TestObject", loc)

	if obj == nil {
		t.Fatal("NewWorldObject() returned nil")
	}

	if obj.ObjectID() != 12345 {
		t.Errorf("ObjectID() = %d, want 12345", obj.ObjectID())
	}

	if obj.Name() != "TestObject" {
		t.Errorf("Name() = %q, want %q", obj.Name(), "TestObject")
	}

	gotLoc := obj.Location()
	if gotLoc != loc {
		t.Errorf("Location() = %+v, want %+v", gotLoc, loc)
	}
}

func TestWorldObject_ObjectID_Immutable(t *testing.T) {
	obj := NewWorldObject(100, "Test", NewLocation(0, 0, 0, 0))

	// ObjectID должен быть immutable — нет сеттера
	id1 := obj.ObjectID()
	id2 := obj.ObjectID()

	if id1 != id2 {
		t.Errorf("ObjectID changed: first=%d, second=%d", id1, id2)
	}

	if id1 != 100 {
		t.Errorf("ObjectID() = %d, want 100", id1)
	}
}

func TestWorldObject_Name(t *testing.T) {
	obj := NewWorldObject(1, "InitialName", NewLocation(0, 0, 0, 0))

	// Проверяем getter
	if obj.Name() != "InitialName" {
		t.Errorf("Name() = %q, want %q", obj.Name(), "InitialName")
	}

	// Проверяем setter
	obj.SetName("UpdatedName")
	if obj.Name() != "UpdatedName" {
		t.Errorf("After SetName, Name() = %q, want %q", obj.Name(), "UpdatedName")
	}

	// Проверяем пустое имя (валидно)
	obj.SetName("")
	if obj.Name() != "" {
		t.Errorf("After SetName empty, Name() = %q, want empty", obj.Name())
	}
}

func TestWorldObject_Location(t *testing.T) {
	initialLoc := NewLocation(100, 200, 300, 1000)
	obj := NewWorldObject(1, "Test", initialLoc)

	// Проверяем getter
	gotLoc := obj.Location()
	if gotLoc != initialLoc {
		t.Errorf("Location() = %+v, want %+v", gotLoc, initialLoc)
	}

	// Проверяем setter
	newLoc := NewLocation(400, 500, 600, 2000)
	obj.SetLocation(newLoc)

	gotLoc = obj.Location()
	if gotLoc != newLoc {
		t.Errorf("After SetLocation, Location() = %+v, want %+v", gotLoc, newLoc)
	}

	// Проверяем что Location() возвращает копию (value type)
	returned := obj.Location()
	modified := returned.WithCoordinates(999, 999, 999)

	// Исходный объект не должен измениться
	if obj.X() == 999 {
		t.Error("Location() did not return a copy - original was mutated")
	}

	// Проверяем что модифицированная копия изменилась
	if modified.X != 999 {
		t.Errorf("Modified copy X = %d, want 999", modified.X)
	}
}

func TestWorldObject_ConvenienceMethods(t *testing.T) {
	loc := NewLocation(111, 222, 333, 4444)
	obj := NewWorldObject(1, "Test", loc)

	// Проверяем convenience methods
	if obj.X() != 111 {
		t.Errorf("X() = %d, want 111", obj.X())
	}
	if obj.Y() != 222 {
		t.Errorf("Y() = %d, want 222", obj.Y())
	}
	if obj.Z() != 333 {
		t.Errorf("Z() = %d, want 333", obj.Z())
	}
	if obj.Heading() != 4444 {
		t.Errorf("Heading() = %d, want 4444", obj.Heading())
	}

	// После SetLocation convenience methods должны вернуть новые значения
	obj.SetLocation(NewLocation(999, 888, 777, 6666))

	if obj.X() != 999 {
		t.Errorf("After SetLocation, X() = %d, want 999", obj.X())
	}
	if obj.Y() != 888 {
		t.Errorf("After SetLocation, Y() = %d, want 888", obj.Y())
	}
	if obj.Z() != 777 {
		t.Errorf("After SetLocation, Z() = %d, want 777", obj.Z())
	}
	if obj.Heading() != 6666 {
		t.Errorf("After SetLocation, Heading() = %d, want 6666", obj.Heading())
	}
}

func TestWorldObject_ConcurrentReads(t *testing.T) {
	obj := NewWorldObject(1, "Test", NewLocation(100, 200, 300, 1000))

	const numReaders = 100
	var wg sync.WaitGroup
	wg.Add(numReaders)

	// Запускаем 100 concurrent readers
	for range numReaders {
		go func() {
			defer wg.Done()

			// Читаем все поля многократно
			for range 1000 {
				_ = obj.ObjectID()
				_ = obj.Name()
				_ = obj.Location()
				_ = obj.X()
				_ = obj.Y()
				_ = obj.Z()
				_ = obj.Heading()
			}
		}()
	}

	wg.Wait()
}

func TestWorldObject_ConcurrentWrites(t *testing.T) {
	obj := NewWorldObject(1, "Test", NewLocation(0, 0, 0, 0))

	const numWriters = 50
	var wg sync.WaitGroup
	wg.Add(numWriters)

	// Запускаем 50 concurrent writers для Name
	for i := range numWriters {
		go func(id int) {
			defer wg.Done()

			for j := range 100 {
				// Каждый writer устанавливает уникальное имя
				obj.SetName("Writer" + string(rune('A'+id)) + string(rune('0'+j%10)))
			}
		}(i)
	}

	wg.Wait()

	// Проверяем что Name() не паникует после concurrent writes
	name := obj.Name()
	if len(name) == 0 {
		t.Error("Name is empty after concurrent writes")
	}
}

func TestWorldObject_ConcurrentLocationUpdates(t *testing.T) {
	obj := NewWorldObject(1, "Test", NewLocation(0, 0, 0, 0))

	const numUpdaters = 50
	var wg sync.WaitGroup
	wg.Add(numUpdaters)

	// Запускаем 50 concurrent updaters для Location
	for i := range numUpdaters {
		go func(id int) {
			defer wg.Done()

			for j := range 100 {
				// Каждый updater устанавливает разные координаты
				x := int32(id*1000 + j)
				y := int32(id*2000 + j)
				z := int32(id*3000 + j)
				heading := uint16(id*100 + j)
				obj.SetLocation(NewLocation(x, y, z, heading))
			}
		}(i)
	}

	wg.Wait()

	// Проверяем что Location() не паникует и возвращает консистентные данные
	loc := obj.Location()
	// Координаты должны быть от одного из writers
	if loc.X < 0 || loc.Y < 0 || loc.Z < 0 {
		t.Errorf("Invalid location after concurrent updates: %+v", loc)
	}
}

func TestWorldObject_MixedReadWrite(t *testing.T) {
	obj := NewWorldObject(1, "Test", NewLocation(100, 200, 300, 1000))

	const numReaders = 50
	const numWriters = 10
	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	// Запускаем readers
	for range numReaders {
		go func() {
			defer wg.Done()

			for range 500 {
				_ = obj.Name()
				_ = obj.Location()
			}
		}()
	}

	// Запускаем writers
	for i := range numWriters {
		go func(id int) {
			defer wg.Done()

			for j := range 100 {
				obj.SetName("Writer" + string(rune('A'+id)))
				obj.SetLocation(NewLocation(int32(id*100+j), int32(id*200+j), int32(id*300+j), uint16(id*10+j)))
			}
		}(i)
	}

	wg.Wait()

	// Финальная проверка консистентности
	name := obj.Name()
	loc := obj.Location()

	if len(name) == 0 {
		t.Error("Name is empty after mixed read/write")
	}
	if loc.X < 0 || loc.Y < 0 || loc.Z < 0 {
		t.Errorf("Invalid location after mixed read/write: %+v", loc)
	}
}

// Benchmark для hot path methods
func BenchmarkWorldObject_Location(b *testing.B) {
	obj := NewWorldObject(1, "Test", NewLocation(100, 200, 300, 1000))

	b.ResetTimer()
	for b.Loop() {
		_ = obj.Location()
	}
}

func BenchmarkWorldObject_X(b *testing.B) {
	obj := NewWorldObject(1, "Test", NewLocation(100, 200, 300, 1000))

	b.ResetTimer()
	for b.Loop() {
		_ = obj.X()
	}
}

func BenchmarkWorldObject_SetLocation(b *testing.B) {
	obj := NewWorldObject(1, "Test", NewLocation(0, 0, 0, 0))
	loc := NewLocation(100, 200, 300, 1000)

	b.ResetTimer()
	for b.Loop() {
		obj.SetLocation(loc)
	}
}
