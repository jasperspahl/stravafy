package sessions

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"stravafy/internal/database"
	"stravafy/internal/templates"
	"strings"
)

var base32RawStrEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

var (
	ErrSessionNotValid       = errors.New("session is not valid")
	ErrNotLoggedIn           = errors.New("user is not logged in")
	ErrUnableToCreateSession = errors.New("unable to create session")
	ErrUnableToFindSession   = errors.New("unable to find session")
)

var (
	sessionCookie = "session_id"
	sessionKey    = "stravafy/internal/sessions/session"
)

type Session interface {
	GetUserId(ctx context.Context) (int64, error)
	GetUser(ctx context.Context) (database.User, error)
	SetUserId(ctx context.Context, userID int64) error
	Logout(ctx context.Context)
}

type session struct {
	sessionID string
	queries   *database.Queries
}

func newSession(ctx *gin.Context, queries *database.Queries) (*session, error) {
	s := &session{
		sessionID: base32RawStrEncoding.EncodeToString(generateRandomKey()),
		queries:   queries,
	}
	clientIP := ctx.ClientIP()
	var userAgent sql.NullString
	value, ok := ctx.Get("User-Agent")
	if ok {
		userAgent.String = value.(string)
		userAgent.Valid = true
	}
	err := s.queries.InsertSession(ctx, database.InsertSessionParams{
		SessionID: s.sessionID,
		UserAgent: userAgent,
		IpAddress: sql.NullString{
			String: clientIP,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, ErrUnableToCreateSession
	}
	ctx.SetCookie(sessionCookie, s.sessionID, 3600, "/", ctx.Request.Host, true, true)
	return s, nil
}

func validSession(c *gin.Context, queries *database.Queries, sessionID string) (*session, error) {
	_, err := queries.GetSession(c, sessionID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return newSession(c, queries)
	}
	err = queries.UpdateSessionLastActivityTime(c, sessionID)
	if err != nil {
		return nil, err
	}
	return &session{
		sessionID: sessionID,
		queries:   queries,
	}, nil
}

func (s *session) GetUser(ctx context.Context) (database.User, error) {
	userID, err := s.GetUserId(ctx)
	if err != nil {
		return database.User{}, ErrNotLoggedIn
	}
	user, _ := s.queries.GetUserById(ctx, userID)
	return user, nil
}

func (s *session) GetUserId(ctx context.Context) (int64, error) {
	userID, err := s.queries.GetUserIdFromSession(ctx, s.sessionID)
	if err != nil {
		log.Printf("Session.GetUserID: %v", err)
		return 0, ErrSessionNotValid
	}
	if !userID.Valid {
		return 0, ErrNotLoggedIn
	}
	return userID.Int64, nil
}

func (s *session) SetUserId(ctx context.Context, userId int64) error {
	return s.queries.UpdateSessionUserId(ctx, database.UpdateSessionUserIdParams{
		UserID:    sql.NullInt64{Int64: userId, Valid: true},
		SessionID: s.sessionID,
	})
}

func (s *session) Logout(ctx context.Context) {
	_ = s.queries.DeleteSession(ctx, s.sessionID)
	s.sessionID = ""
}

func Middleware(q *database.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(sessionCookie)
		var s *session
		if err != nil || sessionID == "" {
			s, err = newSession(c, q)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "", templates.Error(http.StatusInternalServerError, err.Error(), nil))
				return
			}
		} else {
			s, err = validSession(c, q, sessionID)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "", templates.Error(http.StatusInternalServerError, err.Error(), nil))
				return
			}
		}
		log.Print(s.GetUser(c))

		c.Set(sessionKey, s)

		c.Next()

		c.SetCookie(sessionCookie, s.sessionID, 3600, "/", strings.Split(c.Request.Host, ":")[0], false, true)

	}
}

func GetSession(c *gin.Context) (Session, error) {
	value, ok := c.Get(sessionKey)
	if ok {
		s := value.(Session)
		return s, nil
	} else {
		return nil, ErrUnableToFindSession
	}
}

func generateRandomKey() []byte {
	k := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}
