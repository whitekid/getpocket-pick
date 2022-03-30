package pocket

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/pocket-pick/config"
)

func TestGetAuthorizedURL(t *testing.T) {
	api := NewGetPocketAPI(config.ConsumerKey(), "")

	token, url, err := api.AuthorizedURL(context.Background(), "http://127.0.0.1")
	require.NoError(t, err, "error = %v", err)
	require.NotEqual(t, "", token)
	require.NotEqual(t, "", url)
}

func TestAuthorize(t *testing.T) {
	// need to mock web site
}

func TestArticleSearch(t *testing.T) {
	type args struct {
		url string
	}

	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{"url", args{"https://brunch.co.kr/@lunarshore/285"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := NewGetPocketAPI(config.ConsumerKey(), config.AccessToken())
			items, err := api.Articles.Get(context.Background(), WithSearch(tt.args.url))
			require.NoError(t, err)
			require.Equal(t, 1, len(items))

			for _, item := range items {
				require.Equal(t, tt.args.url, item.ResolvedURL)
			}
		})
	}

}

func TestArticleDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	api := NewGetPocketAPI(config.ConsumerKey(), config.AccessToken())

	itemID, err := api.Articles.Add(ctx, "https://news.v.daum.net/v/20220331000726592")
	require.NoError(t, err)
	require.NotEqual(t, "", itemID)

	require.NoError(t, api.Articles.Delete(ctx, itemID))
}
