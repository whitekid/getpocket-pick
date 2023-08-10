package pocket

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/goxp/request"
)

func TestCheckFetchArticle(t *testing.T) {
	type args struct {
		url string
	}

	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{"https://infuture.kr/1688"}, false},
		{"", args{"https://m.biz.chosun.com/svc/article.html?contid=2016012201926"}, false},
		{"", args{"http://blog.naver.com/inno_life/162500428"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := request.Get("https://infuture.kr/1271").
				Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36").
				Do(context.Background())
			if tt.wantErr && err != nil {
				require.Fail(t, "wantErr: %s but got success", tt.wantErr)
				require.NoError(t, resp.Success())
			}
			require.NoError(t, resp.Success())
		})
	}
}
