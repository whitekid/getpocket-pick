package pocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/allegro/bigcache"
	"github.com/stretchr/testify/require"
	"github.com/whitekid/goxp/request"
)

func newTestServer(ctx context.Context) *httptest.Server {
	s := New().(*pocketService)
	e := s.setupRoute()

	ts := httptest.NewServer(e)
	go func() {
		<-ctx.Done()
		ts.Close()
	}()
	return ts
}

func TestSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ts := newTestServer(ctx)

	sess := request.NewSession(nil)

	for i := 0; i < 10; i++ {
		resp, err := sess.Get("%s%s", ts.URL, "/sessions").Do(ctx)
		require.NotEqual(t, 0, len(resp.Cookies()), "cookie must be exists")
		require.NoError(t, err)
		require.True(t, resp.Success(), "status=%d", resp.StatusCode)

		var v string
		require.NoError(t, resp.JSON(&v))
		require.Equal(t, strconv.FormatInt(int64(i), 10), v, "should increase cookie foo")
	}
}

func TestIndex(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ts := newTestServer(ctx)

	// check if redirect to authorize url
	resp, err := request.Get("%s", ts.URL).FollowRedirect(false).Do(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, resp.StatusCode)
	require.True(t, strings.HasPrefix(resp.Header.Get("Location"), "https://getpocket.com/auth/authorize?request_token="), resp.Header.Get("Location"))
}

func TestAuth(t *testing.T) {
	// panic("Not Implemented")
}

func TestCache(t *testing.T) {
	config := bigcache.DefaultConfig(time.Millisecond * 100)
	config.CleanWindow = time.Second

	cache, _ := bigcache.NewBigCache(config)
	require.NoError(t, cache.Set("hello", []byte("world")))
	value, err := cache.Get("hello")
	require.NoError(t, err)
	require.Equal(t, []byte("world"), value)

	// eviction
	{
		time.Sleep(time.Second * 2)
		value, err := cache.Get("hello")
		require.Equal(t, bigcache.ErrEntryNotFound, err)
		require.NotEqual(t, []byte("world"), value)
	}
}
