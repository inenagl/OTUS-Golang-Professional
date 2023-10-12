package memorystorage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestGetEvent(t *testing.T) {
	id := uuid.New()
	nonexistentID := uuid.New()
	event := storage.Event{
		ID:           id,
		Title:        "1 hour later, duration 15 minutes",
		StartDate:    time.Now().Add(time.Hour),
		EndDate:      time.Now().Add(75 * time.Minute),
		Description:  "First event desc",
		UserID:       uuid.New(),
		NotifyBefore: time.Hour,
	}

	s := New()
	s.data[event.ID] = event

	ev, err := s.GetEvent(id)
	require.NoError(t, err)
	require.Equal(t, event, ev)

	_, err = s.GetEvent(nonexistentID)
	require.Error(t, err)
	require.ErrorIs(t, err, storage.ErrEventNotFound)
}

func TestAddEvent(t *testing.T) {
	event := storage.Event{
		Title:        "1 hour later, duration 15 minutes",
		StartDate:    time.Now().Add(time.Hour),
		EndDate:      time.Now().Add(75 * time.Minute),
		Description:  "First event desc",
		UserID:       uuid.New(),
		NotifyBefore: time.Hour,
	}

	s := New()
	require.Equal(t, 0, len(s.data))

	res, err := s.AddEvent(event)
	require.NoError(t, err)
	require.Equal(t, 1, len(s.data))
	event.ID = res.ID
	require.Equal(t, event, s.data[res.ID])
	require.Equal(t, event, res)

	// Повторное добавление такого же ивента приводит к дублированию данных с под новым ID
	res, err = s.AddEvent(event)
	require.NoError(t, err)
	require.Equal(t, 2, len(s.data))
	event2 := event
	event2.ID = res.ID
	require.Equal(t, event2, s.data[res.ID])
	require.Equal(t, event2, res)
}

func TestUpdateEvent(t *testing.T) {
	event := storage.Event{
		ID:           uuid.New(),
		Title:        "1 hour later, duration 15 minutes",
		StartDate:    time.Now().Add(time.Hour),
		EndDate:      time.Now().Add(75 * time.Minute),
		Description:  "First event",
		UserID:       uuid.New(),
		NotifyBefore: time.Hour,
	}

	nonexistentEvent := storage.Event{
		ID:           uuid.New(),
		Title:        "1 day later, duration 1 hour",
		StartDate:    time.Now().Add(24 * time.Hour),
		EndDate:      time.Now().Add(25 * time.Hour),
		Description:  "Nonexistent event",
		UserID:       uuid.New(),
		NotifyBefore: time.Hour,
	}

	s := New()
	s.data[event.ID] = event

	updatedEvent := event
	updatedEvent.StartDate = updatedEvent.StartDate.AddDate(0, 0, 1)
	updatedEvent.EndDate = updatedEvent.EndDate.AddDate(0, 0, 1)
	updatedEvent.Title = "Updated event"
	updatedEvent.NotifyBefore = 0
	err := s.UpdateEvent(updatedEvent)
	require.NoError(t, err)
	require.Equal(t, updatedEvent, s.data[event.ID])

	err = s.UpdateEvent(nonexistentEvent)
	require.Error(t, err)
	require.ErrorIs(t, err, storage.ErrEventNotFound)
}

func TestDeleteEvent(t *testing.T) {
	event := storage.Event{
		ID:           uuid.New(),
		Title:        "1 hour later, duration 15 minutes",
		StartDate:    time.Now().Add(time.Hour),
		EndDate:      time.Now().Add(75 * time.Minute),
		Description:  "First event",
		UserID:       uuid.New(),
		NotifyBefore: time.Hour,
	}

	event2 := storage.Event{
		ID:           uuid.New(),
		Title:        "1 day later, duration 1 hour",
		StartDate:    time.Now().Add(24 * time.Hour),
		EndDate:      time.Now().Add(25 * time.Hour),
		Description:  "Nonexistent event",
		UserID:       uuid.New(),
		NotifyBefore: time.Hour,
	}

	s := New()
	s.data[event.ID] = event
	s.data[event2.ID] = event2
	require.Equal(t, 2, len(s.data))

	err := s.DeleteEvent(event.ID)
	require.NoError(t, err)
	require.Equal(t, 1, len(s.data))
	_, err = s.GetEvent(event.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, storage.ErrEventNotFound)

	ev, err := s.GetEvent(event2.ID)
	require.NoError(t, err)
	require.Equal(t, event2, ev)

	err = s.DeleteEvent(event.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, storage.ErrEventNotFound)
}

func TestGetEvents(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	now := time.Now()

	events := []storage.Event{
		{
			ID:           uuid.New(),
			Title:        "Event 0",
			StartDate:    now,
			EndDate:      now.Add(15 * time.Minute),
			Description:  "User 1, now, duration 15 minutes",
			UserID:       userID1,
			NotifyBefore: time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 1",
			StartDate:    now.Add(time.Hour),
			EndDate:      now.Add(time.Hour).Add(15 * time.Minute),
			Description:  "User 1, 1 hour later, duration 15 minutes",
			UserID:       userID1,
			NotifyBefore: time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 2",
			StartDate:    now.AddDate(0, 0, 1),
			EndDate:      now.AddDate(0, 0, 1).Add(15 * time.Minute),
			Description:  "User 1, 1 day later, duration 15 minutes",
			UserID:       userID1,
			NotifyBefore: time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 3",
			StartDate:    now.AddDate(0, -1, 0),
			EndDate:      now.AddDate(0, -1, 0).AddDate(0, 0, 1),
			Description:  "User 2, last month, duration 1 day",
			UserID:       userID2,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 4",
			StartDate:    now,
			EndDate:      now.AddDate(0, 0, 1),
			Description:  "User 2, now, duration 1 day",
			UserID:       userID2,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 5",
			StartDate:    now.Add(time.Hour),
			EndDate:      now.Add(time.Hour).Add(15 * time.Minute),
			Description:  "User 2, 1 hour later, duration 15 minutes",
			UserID:       userID2,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 6",
			StartDate:    now.AddDate(0, 1, 0),
			EndDate:      now.AddDate(0, 1, 1),
			Description:  "User 2, 1 month later, duration 1 day",
			UserID:       userID2,
			NotifyBefore: 24 * time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 7",
			StartDate:    now.Add(time.Hour),
			EndDate:      now.Add(time.Hour).AddDate(0, 0, 1),
			Description:  "User 3, 1 hour later, duration 1 day",
			UserID:       userID3,
			NotifyBefore: 15 * time.Minute,
		},
	}

	s := New()
	require.Equal(t, 0, len(s.data))

	for _, e := range events {
		s.data[e.ID] = e
	}

	t.Run(
		"Full list", func(t *testing.T) {
			t.Parallel()
			res, err := s.GetEvents([]storage.EventCondition{}, []storage.EventSort{})
			require.NoError(t, err)
			require.ElementsMatch(t, events, res)
		},
	)

	t.Run(
		"Up to now events for user 1 and 2", func(t *testing.T) {
			t.Parallel()
			conds := []storage.EventCondition{
				{Field: storage.EventUserID, Type: storage.TypeIn, Sample: []uuid.UUID{userID1, userID2}},
				{Field: storage.EventStartDate, Type: storage.TypeLessOrEq, Sample: now},
			}
			expected := []storage.Event{events[0], events[3], events[4]}
			res, err := s.GetEvents(conds, []storage.EventSort{})
			require.NoError(t, err)
			require.ElementsMatch(t, expected, res)
		},
	)

	t.Run(
		"Events for user 1 and 2 between now and tomorrow, ordered by enddate DESC and description DESC",
		func(t *testing.T) {
			t.Parallel()
			conds := []storage.EventCondition{
				{Field: storage.EventUserID, Type: storage.TypeIn, Sample: []uuid.UUID{userID1, userID2}},
				{Field: storage.EventStartDate, Type: storage.TypeMore, Sample: now},
				{Field: storage.EventStartDate, Type: storage.TypeLessOrEq, Sample: now.AddDate(0, 0, 1)},
			}
			order := []storage.EventSort{
				{Field: storage.EventEndDate, Direction: storage.DirectionDesc},
				{Field: storage.EventDescription, Direction: storage.DirectionDesc},
			}
			expected := []storage.Event{events[2], events[5], events[1]}
			res, err := s.GetEvents(conds, order)
			require.NoError(t, err)
			require.Equal(t, expected, res)
		},
	)
}

func TestDeleteEvents(t *testing.T) {
	userID := uuid.New()

	now := time.Now()
	events := []storage.Event{
		{
			ID:           uuid.New(),
			Title:        "Event 0",
			StartDate:    now,
			EndDate:      now.Add(15 * time.Minute),
			Description:  "User 1, now, duration 15 minutes",
			UserID:       userID,
			NotifyBefore: time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 1",
			StartDate:    now.AddDate(0, 0, -1).Add(-2 * time.Hour),
			EndDate:      now.AddDate(0, 0, -1).Add(-1 * time.Hour),
			Description:  "User 1, 1 day and 2 hours before, duration 1 hour",
			UserID:       userID,
			NotifyBefore: time.Hour,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 2",
			StartDate:    now.AddDate(0, 0, 1),
			EndDate:      now.AddDate(0, 0, 1).Add(15 * time.Minute),
			Description:  "User 1, 1 day later, duration 15 minutes",
			UserID:       userID,
			NotifyBefore: time.Hour,
		},
	}
	s := New()
	require.Equal(t, 0, len(s.data))

	for _, e := range events {
		s.data[e.ID] = e
	}

	conds := []storage.EventCondition{
		{Field: storage.EventEndDate, Type: storage.TypeLessOrEq, Sample: now.AddDate(0, 0, -1)},
	}
	err := s.DeleteEvents(conds)
	require.NoError(t, err)

	expected := map[uuid.UUID]storage.Event{
		events[0].ID: events[0],
		events[2].ID: events[2],
	}
	require.Equal(t, expected, s.data)
}

func TestSetEventsNotified(t *testing.T) {
	userID := uuid.New()

	now := time.Now()
	events := []storage.Event{
		{
			ID:           uuid.New(),
			Title:        "Event 0",
			StartDate:    now,
			EndDate:      now.Add(15 * time.Minute),
			Description:  "Notified 1 day ago",
			UserID:       userID,
			NotifyBefore: time.Hour,
			NotifiedAt:   now.AddDate(0, 0, -1),
		},
		{
			ID:           uuid.New(),
			Title:        "Event 1",
			StartDate:    now.AddDate(0, 0, -1).Add(-2 * time.Hour),
			EndDate:      now.AddDate(0, 0, -1).Add(-1 * time.Hour),
			Description:  "Notified 2 day ago",
			UserID:       userID,
			NotifyBefore: time.Hour,
			NotifiedAt:   now.AddDate(0, 0, -2),
		},
		{
			ID:           uuid.New(),
			Title:        "Event 2",
			StartDate:    now.AddDate(0, 0, 1),
			EndDate:      now.AddDate(0, 0, 1).Add(15 * time.Minute),
			Description:  "Notified 3 day ago",
			UserID:       userID,
			NotifyBefore: time.Hour,
			NotifiedAt:   now.AddDate(0, 0, -3),
		},
	}
	s := New()
	require.Equal(t, 0, len(s.data))

	for _, e := range events {
		s.data[e.ID] = e
	}

	err := s.SetEventsNotified([]uuid.UUID{events[1].ID, events[2].ID}, now)
	require.NoError(t, err)

	events[1].NotifiedAt = now
	events[2].NotifiedAt = now
	expected := map[uuid.UUID]storage.Event{
		events[0].ID: events[0],
		events[1].ID: events[1],
		events[2].ID: events[2],
	}
	require.Equal(t, expected, s.data)
}

func TestNotificationNeededEvents(t *testing.T) {
	userID := uuid.New()

	now := time.Now()
	events := []storage.Event{
		{
			ID:           uuid.New(),
			Title:        "Event 0",
			StartDate:    now,
			EndDate:      now.Add(15 * time.Minute),
			Description:  "Start now, notify 1 hour, notified day before",
			UserID:       userID,
			NotifyBefore: time.Hour,
			NotifiedAt:   now.AddDate(0, 0, -1),
		},
		{
			ID:           uuid.New(),
			Title:        "Event 1",
			StartDate:    now.AddDate(0, 0, 1),
			EndDate:      now.AddDate(0, 0, 1).Add(1 * time.Hour),
			Description:  "Start tomorrow, notify 1 minute before, notified now",
			UserID:       userID,
			NotifyBefore: 24*time.Hour + 1*time.Minute,
			NotifiedAt:   now,
		},
		{
			ID:           uuid.New(),
			Title:        "Event 2",
			StartDate:    now.AddDate(0, 0, 1),
			EndDate:      now.AddDate(0, 0, 1).Add(1 * time.Hour),
			Description:  "Start tomorrow, notify 1 minute before, notified yesterday",
			UserID:       userID,
			NotifyBefore: 24*time.Hour + 1*time.Minute,
			NotifiedAt:   now.AddDate(0, 0, -1),
		},
	}
	s := New()
	require.Equal(t, 0, len(s.data))

	for _, e := range events {
		s.data[e.ID] = e
	}

	res, err := s.NotificationNeededEvents(now)
	require.NoError(t, err)

	expected := []storage.Event{events[0], events[2]}
	require.ElementsMatch(t, expected, res)
}
