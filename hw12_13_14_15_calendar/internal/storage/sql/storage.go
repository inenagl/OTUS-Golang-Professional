package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
)

var fieldsMap = map[storage.EventField]string{
	storage.EventID:           "id",
	storage.EventTitle:        "title",
	storage.EventStartDate:    "start_date",
	storage.EventEndDate:      "end_date",
	storage.EventDescription:  "description",
	storage.EventUserID:       "user_id",
	storage.EventNotifyBefore: "notify_before",
	storage.EventNotifiedAt:   "notified_at",
}

type Storage struct {
	dsn     string
	db      *sqlx.DB
	timeout time.Duration
}

func New(host string, port int, dbname, user, password, sslmode string, timeout time.Duration) *Storage {
	portStr := ""
	if port != 0 {
		portStr = ":" + strconv.Itoa(port)
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s%s/%s?sslmode=%s", user, password, host, portStr, dbname, sslmode)

	return &Storage{
		dsn:     dsn,
		timeout: timeout,
	}
}

func (s Storage) createTimeoutCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), s.timeout) //nolint: govet
	return ctx
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, "pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	s.db = db

	return nil
}

func (s *Storage) Ping() error {
	if s.db == nil {
		return s.Connect(s.createTimeoutCtx())
	}

	if err := s.db.PingContext(s.createTimeoutCtx()); err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}

	return nil
}

func (s *Storage) AddEvent(event storage.Event) (storage.Event, error) {
	if err := s.Ping(); err != nil {
		return storage.Event{}, err
	}

	query, args, err := sqlx.Named(
		`INSERT INTO Events 
    	(title, description, start_date, end_date, user_id, notify_before, notified_at)
        VALUES (:title, :description, :start_date, :end_date, :user_id, :notify_before, :notified_at) RETURNING id`,
		event,
	)
	if err != nil {
		return storage.Event{}, err
	}
	query = s.db.Rebind(query)

	var id uuid.UUID
	err = s.db.GetContext(s.createTimeoutCtx(), &id, query, args...)
	if err != nil {
		return storage.Event{}, err
	}

	res, err := s.GetEvent(id)
	if err != nil {
		return storage.Event{}, err
	}

	return res, nil
}

func (s *Storage) UpdateEvent(event storage.Event) error {
	if err := s.Ping(); err != nil {
		return err
	}

	_, err := s.db.NamedExecContext(
		s.createTimeoutCtx(),
		`UPDATE Events SET
                  title = :title,
                  description = :description,
                  start_date = :start_date,
                  end_date = :end_date,
                  user_id = :user_id,
                  notify_before = :notify_before,
                  notified_at = :notified_at
            WHERE id=:id`,
		event,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) DeleteEvent(id uuid.UUID) error {
	if err := s.Ping(); err != nil {
		return err
	}

	_, err := s.db.ExecContext(s.createTimeoutCtx(), "DELETE FROM Events WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetEvent(id uuid.UUID) (storage.Event, error) {
	if err := s.Ping(); err != nil {
		return storage.Event{}, err
	}

	event := storage.Event{}
	err := s.db.GetContext(
		s.createTimeoutCtx(),
		&event,
		`
SELECT id, title, description, start_date, end_date, user_id, notify_before, notified_at
FROM Events WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			err = storage.ErrEventNotFound
		}
		return storage.Event{}, err
	}

	return event, nil
}

func (s *Storage) DeleteEvents(filter []storage.EventCondition) (int64, error) {
	if len(filter) == 0 {
		return 0, errors.New("delete: filter required")
	}
	if err := s.Ping(); err != nil {
		return 0, err
	}

	args := make([]interface{}, 0, len(filter))
	where, args := getWhere(filter, args, 1)
	query := fmt.Sprintf("DELETE FROM Events WHERE %s", where)

	res, err := s.db.ExecContext(s.createTimeoutCtx(), query, args...)
	if err != nil {
		return 0, err
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return cnt, nil
}

func (s *Storage) SetEventsNotified(ids []uuid.UUID, notified time.Time) error {
	if len(ids) == 0 {
		return errors.New("set notified: ids required")
	}
	if err := s.Ping(); err != nil {
		return err
	}

	args := make([]interface{}, 1, len(ids)+1)
	args[0] = notified

	filter := []storage.EventCondition{
		{Field: storage.EventID, Type: storage.TypeIn, Sample: ids},
	}
	where, args := getWhere(filter, args, 2)
	query := fmt.Sprintf("UPDATE Events SET notified_at = $1 WHERE %s", where)

	_, err := s.db.ExecContext(s.createTimeoutCtx(), query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) NotificationNeededEvents(t time.Time) ([]storage.Event, error) {
	if err := s.Ping(); err != nil {
		return []storage.Event{}, err
	}

	query := `
SELECT id, title, description, start_date, end_date, user_id, notify_before, notified_at 
FROM Events 
WHERE 
	start_date - make_interval(secs => notify_before/1000000000) < $1
	AND notified_at < start_date - make_interval(secs => notify_before/1000000000)`
	args := []interface{}{t}

	var events []storage.Event
	err := s.db.SelectContext(s.createTimeoutCtx(), &events, query, args...)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Storage) GetEvents(filter []storage.EventCondition, sort []storage.EventSort) ([]storage.Event, error) {
	if err := s.Ping(); err != nil {
		return []storage.Event{}, err
	}

	qb := strings.Builder{}
	qb.WriteString("SELECT id, title, description, start_date, end_date, user_id, notify_before, notified_at FROM Events")

	args := make([]interface{}, 0, len(filter))
	if len(filter) > 0 {
		where, a := getWhere(filter, args, 1)
		qb.WriteString(" WHERE ")
		qb.WriteString(where)
		args = a
	}

	if len(sort) > 0 {
		qb.WriteString(" ORDER BY ")
		qb.WriteString(getSort(sort))
	}

	var events []storage.Event
	err := s.db.SelectContext(s.createTimeoutCtx(), &events, qb.String(), args...)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func getSort(sorts []storage.EventSort) string {
	s := make([]string, len(sorts))
	i := 0
	for _, v := range sorts {
		s[i] = fieldsMap[v.Field] + " " + string(v.Direction)
		i++
	}

	return strings.Join(s, ", ")
}

func getWhere(filter []storage.EventCondition, args []interface{}, startPos int) (string, []interface{}) {
	wheres := make([]string, 0, len(filter))
	i := startPos
	for _, v := range filter {
		if v.Type == storage.TypeIn || v.Type == storage.TypeNotIn {
			var inStr []string
			switch t := v.Sample.(type) {
			case []string:
				for _, val := range t {
					inStr = append(inStr, "$"+strconv.Itoa(i))
					args = append(args, val)
					i++
				}
			case []int:
				for _, val := range t {
					inStr = append(inStr, "$"+strconv.Itoa(i))
					args = append(args, val)
					i++
				}
			case []uuid.UUID:
				for _, val := range t {
					inStr = append(inStr, "$"+strconv.Itoa(i))
					args = append(args, val.String())
					i++
				}
			default:
				for _, val := range v.Sample.([]interface{}) {
					inStr = append(inStr, "$"+strconv.Itoa(i))
					args = append(args, val)
					i++
				}
			}
			wheres = append(wheres, fmt.Sprintf("%s %s (%s)", fieldsMap[v.Field], v.Type, strings.Join(inStr, ",")))
		} else {
			wheres = append(wheres, fmt.Sprintf("%s %s $%d", fieldsMap[v.Field], v.Type, i))
			args = append(args, v.Sample)
			i++
		}
	}

	return strings.Join(wheres, " AND "), args
}
