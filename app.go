package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/whitekid/echox"
	"github.com/whitekid/getpocket"
	"github.com/whitekid/goxp/fx"
	"github.com/whitekid/goxp/log"
	"github.com/whitekid/goxp/service"

	"pocket-pick/config"
	"pocket-pick/pkg/cache"
)

const (
	keyRequestToken = "REQUEST_TOKEN"
	keyAccessToken  = "ACCESS_TOKEN"
)

// New return pocket-pick service object
// implements service interface
func New(ctx context.Context) service.Interface {
	rootURL := config.RootURL()
	if rootURL == "" {
		panic("ROOT_URL required")
	}

	return &pocketService{
		cache:   cache.NewBigCache(ctx),
		rootURL: rootURL,
	}
}

type pocketService struct {
	rootURL string
	cache   cache.Interface // for api cache
}

// Serve serve the main service
func (s *pocketService) Serve(ctx context.Context) error {
	e := s.setupRoute()

	go func() {
		<-ctx.Done()
		if err := e.Shutdown(context.Background()); err != nil {
			log.Fatalf("%s", err)
		}
	}()

	if err := e.Start(config.BindAddr()); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

type Context struct {
	echo.Context

	sess *sessions.Session
}

type ContextFactory struct{}

func (f *ContextFactory) ContextIs(c echo.Context) bool {
	_, ok := c.(*Context)
	return ok
}

func (f *ContextFactory) NewContext(c echo.Context) echo.Context {
	return &Context{Context: c}
}

func (s *pocketService) setupRoute() *echox.Echo {
	e := echox.New()
	e.Use(echox.CustomContext(&ContextFactory{}))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))),
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				cc := c.(*Context)

				if cc.sess == nil {
					sess, _ := session.Get("pocket-pick-session", cc)
					sess.Options = &sessions.Options{
						Path:     "/",
						MaxAge:   int(config.CookieTimeout().Seconds()),
						HttpOnly: true,
					}
					cc.sess = sess
				}

				return next(cc)
			}
		})

	e.GET("/", s.handleGetIndex)
	e.GET("/auth", s.handleGetAuth)
	e.GET("/article/:item_id", s.handleGetArticle) // TODO 원래는 DELETE로 해야하는데, 귀찮아서..
	e.GET("/sessions", s.handleGetSession)

	return e
}

func (s *pocketService) session(c echo.Context) *sessions.Session { return c.(*Context).sess }

func (s *pocketService) handleGetSession(c echo.Context) error {
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

func (s *pocketService) handleGetIndex(c echo.Context) error {
	sess := s.session(c)
	ctx := c.Request().Context()

	// if not token, try to authorize
	if _, exists := sess.Values[keyRequestToken]; !exists {
		requestToken, authorizedURL, err := getpocket.New(config.ConsumerKey(), "").
			AuthorizedURL(ctx, fmt.Sprintf("%s/auth", s.rootURL))
		if err != nil {
			return errors.Wrapf(err, "authorize failed")
		}

		sess.Values[keyRequestToken] = requestToken
		log.Infof("save requestToken to session: %s", requestToken)
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, authorizedURL)
	}

	// aleardy has a token, redirect to root to get article
	if _, exists := sess.Values[keyAccessToken]; !exists {
		delete(sess.Values, keyRequestToken)
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	accessToken := sess.Values[keyAccessToken].(string)
	log.Debugf("accessToken acquired, get random favorite pick: %s", accessToken)

	key := fmt.Sprintf("%s/favorites", accessToken)
	api := getpocket.New(config.ConsumerKey(), accessToken)

	data, err := s.cache.Get(ctx, key)
	var articleList map[string]*getpocket.Article
	if err != nil {
		if err != cache.ErrNotExists {
			return err
		}
		var err error
		articleList, err = api.Articles().Get().Favorite(getpocket.Favorited).Do(ctx)
		if err != nil {
			return errors.Wrap(err, "get favorite artcles failed")
		}

		// write to cache
		buf, err := json.Marshal(articleList)
		if err != nil {
			return errors.Wrap(err, "json encode failed")
		}
		s.cache.Set(ctx, key, buf, cache.WithExpire(config.CacheEvictionTimeout()))
	} else {
		log.Debug("load articles from cache")

		articleList = make(map[string]*getpocket.Article)
		buf := bytes.NewBuffer(data)
		if err := json.NewDecoder(buf).Decode(&articleList); err != nil {
			return errors.Wrap(err, "json decode failed")
		}
	}

	log.Debugf("you have %d articles", len(articleList))

	// random pick from articles
	_, article := fx.SampleMap(articleList)
	log.Debugf("article: %+v", article)

	url := fmt.Sprintf("https://getpocket.com/read/%s", article.ItemID)

	return c.Redirect(http.StatusFound, url)
}

func (s *pocketService) handleGetAuth(c echo.Context) (err error) {
	sess := s.session(c)

	if _, exists := sess.Values[keyRequestToken]; !exists {
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	requestToken := sess.Values[keyRequestToken].(string)
	if _, exists := sess.Values[keyAccessToken]; !exists {
		accessToken, _, err := getpocket.New(config.ConsumerKey(), "").NewAccessToken(c.Request().Context(), requestToken)
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

func (s *pocketService) requireAccessToken(c echo.Context, token *string) error {
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
func (s *pocketService) handleGetArticle(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ItemID missed")
	}

	var accessToken string

	if err := s.requireAccessToken(c, &accessToken); err != nil {
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	if _, err := getpocket.New(config.ConsumerKey(), accessToken).Modify().Delete(itemID).Do(c.Request().Context()); err != nil {
		log.Errorf("failed: %s", err)
		return err
	}

	return nil
}
