package model

import "sync"

// WorldObject — базовый класс для всех игровых объектов в мире.
// Все объекты имеют ObjectID, Name и Location.
// Phase 5.6: Added Data field for Player/Npc reference.
type WorldObject struct {
	objectID uint32
	name     string
	location Location
	Data     any // Phase 5.6: Player or Npc reference

	mu sync.RWMutex
}

// NewWorldObject создаёт новый объект в игровом мире.
func NewWorldObject(objectID uint32, name string, loc Location) *WorldObject {
	return &WorldObject{
		objectID: objectID,
		name:     name,
		location: loc,
	}
}

// ObjectID возвращает уникальный ID объекта (immutable после создания).
func (w *WorldObject) ObjectID() uint32 {
	return w.objectID
}

// Name возвращает имя объекта.
func (w *WorldObject) Name() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.name
}

// SetName устанавливает имя объекта.
func (w *WorldObject) SetName(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.name = name
}

// Location возвращает копию координат объекта (value type).
func (w *WorldObject) Location() Location {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.location
}

// SetLocation устанавливает новые координаты объекта.
func (w *WorldObject) SetLocation(loc Location) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.location = loc
}

// X возвращает координату X (convenience method для hot path).
func (w *WorldObject) X() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.location.X
}

// Y возвращает координату Y (convenience method для hot path).
func (w *WorldObject) Y() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.location.Y
}

// Z возвращает координату Z (convenience method для hot path).
func (w *WorldObject) Z() int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.location.Z
}

// Heading возвращает направление (convenience method для hot path).
func (w *WorldObject) Heading() uint16 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.location.Heading
}
