package memorystorage

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	mu   sync.RWMutex
	data map[uuid.UUID]storage.Event
}

func New() *Storage {
	return &Storage{data: make(map[uuid.UUID]storage.Event)}
}

func (s *Storage) AddEvent(event storage.Event) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		event.ID = uuid.New()
		_, ok := s.data[event.ID]
		if !ok {
			break
		}
	}

	s.data[event.ID] = event
	return event.ID, nil
}

func (s *Storage) UpdateEvent(event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[event.ID]
	if !ok {
		return fmt.Errorf("%w: ID = %s", storage.ErrEventNotFound, event.ID)
	}
	s.data[event.ID] = event
	return nil
}

func (s *Storage) DeleteEvent(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data[id]
	if !ok {
		return fmt.Errorf("%w: ID = %s", storage.ErrEventNotFound, id)
	}
	delete(s.data, id)
	return nil
}

func (s *Storage) GetEvent(id uuid.UUID) (storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.data[id]
	if !ok {
		return storage.Event{}, fmt.Errorf("%w: ID = %s", storage.ErrEventNotFound, id)
	}
	return e, nil
}

func (s *Storage) GetEvents(search []storage.EventCondition, order []storage.EventSort) ([]storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]storage.Event, 0, len(s.data))
	for _, v := range s.data {
		b, err := searchEvent(v, search)
		if err != nil {
			return []storage.Event{}, err
		}
		if b {
			result = append(result, v)
		}
	}

	if len(order) != 0 {
		lessComp := func(a interface{}, b interface{}) bool {
			switch a := a.(type) {
			case int:
				return a < b.(int)
			case string:
				return strings.Compare(a, b.(string)) == -1
			case time.Time:
				return a.Unix() < b.(time.Time).Unix()
			}
			return false
		}
		sort.Slice(result, func(i, j int) bool {
			var isLess bool
			for _, ord := range order {
				if result[i].GetFieldValue(ord.Field) != result[j].GetFieldValue(ord.Field) {
					isLess = lessComp(result[i].GetFieldValue(ord.Field), result[j].GetFieldValue(ord.Field))

					if ord.Direction == storage.DirectionAsc {
						return isLess
					}
					return !isLess
				}
			}
			return false
		})
	}

	return result, nil
}

func searchEvent(event storage.Event, search []storage.EventCondition) (bool, error) {
	for _, s := range search {
		b, err := applyCondition(event, s)
		if err != nil {
			return false, err
		}
		if !b {
			return false, nil
		}
	}
	return true, nil
}

func applyCondition(event storage.Event, cond storage.EventCondition) (bool, error) {
	val := event.GetFieldValue(cond.Field)

	// Тут бы со стратегиями и т.п. но очень не хватает времени это всё закодить
	switch cond.Type {
	case storage.TypeEq:
		return val == cond.Sample, nil
	case storage.TypeNotEq:
		return val != cond.Sample, nil
	case storage.TypeLess:
		return lessCmp(val, cond.Sample)
	case storage.TypeLessOrEq:
		return lessOrEqCmp(val, cond.Sample)
	case storage.TypeMore:
		return moreCmp(val, cond.Sample)
	case storage.TypeMoreOrEq:
		return moreOrEqCmp(val, cond.Sample)
	case storage.TypeIn:
		return inCmp(val, cond.Sample)
	case storage.TypeNotIn:
		res, err := inCmp(val, cond.Sample)
		if err != nil {
			return false, err
		}
		return !res, nil
	default:
		return false, fmt.Errorf("%w: '%s'", storage.ErrUnknownCondition, cond.Type)
	}
}

func lessCmp(val, sample interface{}) (bool, error) {
	switch v := val.(type) {
	case int:
		return v < sample.(int), nil
	case time.Time:
		return v.Unix() < sample.(time.Time).Unix(), nil
	default:
		return false, fmt.Errorf("%w: %T", storage.ErrIncomparableType, val)
	}
}

func lessOrEqCmp(val, sample interface{}) (bool, error) {
	switch v := val.(type) {
	case int:
		return v <= sample.(int), nil
	case time.Time:
		return v.Unix() <= sample.(time.Time).Unix(), nil
	default:
		return false, fmt.Errorf("%w: %T", storage.ErrIncomparableType, val)
	}
}

func moreCmp(val, sample interface{}) (bool, error) {
	switch v := val.(type) {
	case int:
		return v > sample.(int), nil
	case time.Time:
		return v.Unix() > sample.(time.Time).Unix(), nil
	default:
		return false, fmt.Errorf("%w: %T", storage.ErrIncomparableType, val)
	}
}

func moreOrEqCmp(val, sample interface{}) (bool, error) {
	switch v := val.(type) {
	case int:
		return v >= sample.(int), nil
	case time.Time:
		return v.Unix() >= sample.(time.Time).Unix(), nil
	default:
		return false, fmt.Errorf("%w: %T", storage.ErrIncomparableType, val)
	}
}

func inCmp(val, sample interface{}) (bool, error) {
	switch s := sample.(type) {
	case []int:
		for _, v := range s {
			if val == v {
				return true, nil
			}
		}
	case []uuid.UUID:
		for _, v := range s {
			if val == v {
				return true, nil
			}
		}
	case []string:
		for _, v := range s {
			if val == v {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("%w: %T", storage.ErrIncomparableType, sample)
	}
	return false, nil
}
