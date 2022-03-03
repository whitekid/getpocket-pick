package pocket

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/allegro/bigcache"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/whitekid/go-utils/logging"
	"github.com/whitekid/pocket-pick/pkg/config"
)

const (
	keyRequestToken = "REQUEST_TOKEN"
	keyAccessToken  = "ACCESS_TOKEN"
)

// New implements service interface
func New() *Service {
	rootURL := os.Getenv("ROOT_URL")
	if rootURL == "" {
		panic("ROOT_URL required")
	}

	config := bigcache.DefaultConfig(time.Hour)
	config.CleanWindow = time.Minute

	cache, _ := bigcache.NewBigCache(config)

	return &Service{
		cache:   cache,
		rootURL: rootURL,
	}
}

// Service the main service
type Service struct {
	rootURL string
	cache   *bigcache.BigCache // for api cache
}

// Serve serve the main service
func (s *Service) Serve(ctx context.Context, args ...string) error {
	e := s.setupRoute()

	return e.Start(config.BindAddr())
}

func (s *Service) setupRoute() *echo.Echo {
	e := echo.New()

	loggerConfig := middleware.DefaultLoggerConfig
	e.Use(middleware.LoggerWithConfig(loggerConfig))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.GET("/", s.handleGetIndex)
	e.GET("/auth", s.handleGetAuth)
	e.GET("/article/:item_id", s.handleGetArticle) // TODO 원래는 DELETE로 해야하는데, 귀찮아서..
	e.GET("/sessions", s.handleGetSession)

	return e
}

func (s *Service) session(c echo.Context) *sessions.Session {
	sess, _ := session.Get("pocket-pick-session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
	}

	return sess
}

func (s *Service) handleGetSession(c echo.Context) error {
	sess := s.session(c)
	if sess.Values["foo"] == nil {
		sess.Values["foo"] = "0"
	} else {
		v, err := strconv.Atoi(sess.Values["foo"].(string))
		if err != nil {
			v = 0
		}
		sess.Values["foo"] = strconv.FormatInt(int64(v)+1, 10)
	}
	sess.Save(c.Request(), c.Response())

	if err := c.JSON(http.StatusOK, sess.Values["foo"]); err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (s *Service) handleGetIndex(c echo.Context) error {
	sess := s.session(c)

	// if not token, try to authorize
	if _, exists := sess.Values[keyRequestToken]; !exists {
		requestToken, authorizedURL, err := NewGetPocketAPI(os.Getenv("CONSUMER_KEY"), "").AuthorizedURL(fmt.Sprintf("%s/auth", s.rootURL))
		if err != nil {
			return err
		}

		sess.Values[keyRequestToken] = requestToken
		log.Infof("save requestToken to session: %s", requestToken)
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, authorizedURL)
	}

	if _, exists := sess.Values[keyAccessToken]; !exists {
		delete(sess.Values, keyRequestToken)
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	accessToken := sess.Values[keyAccessToken].(string)
	log.Debugf("accessToken acquired, get random favorite pick: %s", accessToken)
	url, err := getRandomPickURL(s.cache, accessToken)
	if err != nil {
		log.Errorf("error: %s", err)
		return err
	}

	log.Infof("move to %s", url)
	return c.Redirect(http.StatusFound, url)
}

func (s *Service) handleGetAuth(c echo.Context) (err error) {
	sess := s.session(c)

	if _, exists := sess.Values[keyRequestToken]; !exists {
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	requestToken := sess.Values[keyRequestToken].(string)
	if _, exists := sess.Values[keyAccessToken]; !exists {
		accessToken, _, err := NewGetPocketAPI(os.Getenv("CONSUMER_KEY"), "").AccessToken(requestToken)
		if err != nil {
			log.Errorf("fail to get access token: %s", err)
			return err
		}

		if accessToken == "" {
			delete(sess.Values, keyAccessToken)
			sess.Save(c.Request(), c.Response())

			return c.Redirect(http.StatusFound, s.rootURL)
		}

		log.Debugf("get accessToken %s", accessToken)
		sess.Values[keyAccessToken] = accessToken
		sess.Save(c.Request(), c.Response())
	}

	log.Debug("redirect to root to read a item")
	return c.Redirect(http.StatusFound, s.rootURL)
}

func (s *Service) requireAccessToken(c echo.Context, token *string) error {
	sess := s.session(c)

	if _, exists := sess.Values[keyAccessToken]; !exists {
		delete(sess.Values, keyRequestToken)
		sess.Save(c.Request(), c.Response())
		return fmt.Errorf("access token not found")
	}

	*token = sess.Values[keyAccessToken].(string)
	return nil
}

// remove given article
func (s *Service) handleGetArticle(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return c.String(http.StatusBadRequest, "ItemID missed")
	}

	var accessToken string

	if err := s.requireAccessToken(c, &accessToken); err != nil {
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	if err := NewGetPocketAPI(os.Getenv("CONSUMER_KEY"), accessToken).Articles.Delete(itemID); err != nil {
		log.Errorf("failed: %s", err)
		return err
	}

	return nil
}
