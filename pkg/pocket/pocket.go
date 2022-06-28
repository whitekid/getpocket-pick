package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/pkg/errors"
	"github.com/whitekid/goxp/log"
	"github.com/whitekid/goxp/request"
)

// Client get pocket api client
// please refer https://getpocket.com/developer/docs/overview
type Client struct {
	consumerKey string
	accessToken string
	sess        request.Interface // common sessions

	// API interfaces
	Articles *ArticlesAPI
}

// New create GetPocket API
func New(consumerKey, accessToken string) *Client {
	api := &Client{
		consumerKey: consumerKey,
		accessToken: accessToken,
		sess:        request.NewSession(nil),
	}

	api.Articles = &ArticlesAPI{pocket: api}

	return api
}

// Article pocket article, see https://getpocket.com/developer/docs/v3/retrieve
type Article struct {
	ItemID        string `json:"item_id"`
	ResolvedID    string `json:"resolved_id"`
	GivenURL      string `json:"given_url"`
	GivelTitle    string `json:"given_title"`
	Favorite      string `json:"favorite"`
	Status        string `json:"status"`
	ResolvedTitle string `json:"resolved_title"`
	ResolvedURL   string `json:"resolved_url"`
	Excerpt       string `json:"excerpt"`
	IsArticle     string `json:"is_article"`
	HasVideo      string `json:"has_video"`
	HasImage      string `json:"has_image"`
	WordCount     string `json:"word_count"`
	Images        map[string]struct {
		ItemID  string `json:"item_id"`
		ImageID string `json:"image_id"`
		Src     string `json:"src"`
		Width   string `json:"width"`
		Height  string `json:"height"`
		Credit  string `json:"credit"`
		Caption string `json:"caption"`
	} `json:"images"`
	Videos map[string]struct {
		ItemID  string `json:"item_id"`
		VideoID string `json:"video_id"`
		Src     string `json:"src"`
		Width   string `json:"width"`
		Height  string `json:"height"`
		Type    string `json:"type"`
		Vid     string `json:"vid"`
	} `json:"videos"`
}

func (g *Client) success(resp *request.Response) error {
	if resp.Success() {
		return nil
	}

	message := resp.Header.Get("x-error")
	code := resp.Header.Get("x-error-code")
	return fmt.Errorf("error with status: %d, error=%s, code=%s", resp.StatusCode, message, code)
}

// AuthorizedURL get authorizedURL
func (g *Client) AuthorizedURL(ctx context.Context, redirectURI string) (string, string, error) {
	resp, err := request.Post("https://getpocket.com/v3/oauth/request").
		Header("X-Accept", "application/json").
		JSON(map[string]string{
			"consumer_key": g.consumerKey,
			"redirect_uri": redirectURI,
		}).Do(ctx)

	if err != nil {
		return "", "", errors.Wrap(err, "authorized request failed")
	}

	if err := g.success(resp); err != nil {
		return "", "", errors.Wrap(err, "AutorizedURL failed")
	}

	var response struct {
		Code string `json:"code"`
	}

	if err := resp.JSON(&response); err != nil {
		return "", "", errors.Wrap(err, "fail to parse json")
	}
	defer resp.Body.Close()

	return response.Code, fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", response.Code, redirectURI), nil
}

// NewAccessToken get accessToken, username from requestToken using oauth
func (g *Client) NewAccessToken(ctx context.Context, requestToken string) (string, string, error) {
	log.Debugf("getAccessToken with %s", requestToken)

	resp, err := g.sess.Post("https://getpocket.com/v3/oauth/authorize").
		Header("X-Accept", "application/json").
		JSON(map[string]string{
			"consumer_key": g.consumerKey,
			"code":         requestToken,
		}).Do(ctx)
	if err != nil {
		return "", "", err
	}

	if err := g.success(resp); err != nil {
		return "", "", fmt.Errorf("failed with status: %d", resp.StatusCode)
	}

	var response struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}

	defer resp.Body.Close()
	if err := resp.JSON(&response); err != nil {
		return "", "", err
	}

	return response.AccessToken, response.Username, nil
}

// ArticlesAPI ...
type ArticlesAPI struct {
	pocket *Client
}

const (
	UnFavorited = 1 // only return un-favorited items
	Favorited   = 2 // only return favorited items
)

// ArticleGetResponse ...
type ArticleGetResponse struct {
	Status int                  `json:"status"`
	List   *map[string]*Article `json:"list"`
}

// Get Retrieving a User's Pocket Data
func (a *ArticlesAPI) Get() *ArticleGetRequest {
	return &ArticleGetRequest{
		api: a,
	}
}

type ArticleGetRequest struct {
	api *ArticlesAPI

	search   string
	domain   string
	favorite int
}

func (r *ArticleGetRequest) Search(search string) *ArticleGetRequest {
	r.search = search
	return r
}

func (r *ArticleGetRequest) Domain(domain string) *ArticleGetRequest {
	r.domain = domain
	return r
}

func (r *ArticleGetRequest) Favorite(favorite int) *ArticleGetRequest {
	r.favorite = favorite
	return r
}

func (r *ArticleGetRequest) Do(ctx context.Context) (map[string]*Article, error) {
	params := map[string]interface{}{
		"consumer_key": r.api.pocket.consumerKey,
		"access_token": r.api.pocket.accessToken,
		"state":        "all",
		"detailType":   "simple",
	}

	if r.favorite != 0 {
		params["favorite"] = strconv.FormatInt(int64(r.favorite-1), 10)
	}

	if r.search != "" {
		params["search"] = r.search
	}

	if r.domain != "" {
		params["domain"] = r.domain
	}

	resp, err := r.api.pocket.sess.
		Post("https://getpocket.com/v3/get").
		JSON(params).Do(ctx)
	if err != nil {
		return nil, err
	}

	if err := r.api.pocket.success(resp); err != nil {
		return nil, errors.Wrapf(err, "Get()")
	}

	var buffer bytes.Buffer
	var buf1 bytes.Buffer
	io.Copy(&buffer, resp.Body)
	defer resp.Body.Close()

	//
	tee := io.TeeReader(&buffer, &buf1)

	// return empty list if there is no items searched
	var emptyResponse struct {
		List []string `json:"list"`
	}
	if err := json.NewDecoder(tee).Decode(&emptyResponse); err == nil {
		return nil, err
	}

	var response ArticleGetResponse
	if err := json.NewDecoder(&buf1).Decode(&response); err != nil {
		errors.Wrap(err, "JSONDecode")
		return nil, err
	}

	return *response.List, nil
}

type actionParam struct {
	Action string `json:"action"`
	ItemID string `json:"item_id"`
	Time   string `json:"time,omitempty"`
}

type actionResults struct {
	ActionResults []bool `json:"action_results"`
	Status        int    `json:"status"`
}

func (a *ArticlesAPI) sendAction(ctx context.Context, actions []actionParam) (*actionResults, error) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&actions)

	log.Debugf("actions: %+v", actions)
	resp, err := a.pocket.sess.Post("https://getpocket.com/v3/send").
		Form("consumer_key", a.pocket.consumerKey).
		Form("access_token", a.pocket.accessToken).
		Form("actions", buf.String()).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	if err := a.pocket.success(resp); err != nil {
		return nil, errors.Wrap(err, "sendAction()")
	}

	var response actionResults

	defer resp.Body.Close()
	if err := resp.JSON(&response); err != nil {
		return nil, errors.Wrapf(err, "decode response")
	}
	log.Debugf("resp: %+v", response)

	if !response.ActionResults[0] {
		return nil, fmt.Errorf("delete failed: %v, %d", response.ActionResults[0], response.Status)
	}

	return &response, nil
}

func (a *ArticlesAPI) Add(ctx context.Context, url string) (itemID string, err error) {
	resp, err := a.pocket.sess.Post("https://getpocket.com/v3/add").
		JSON(map[string]string{
			"url":          url,
			"consumer_key": a.pocket.consumerKey,
			"access_token": a.pocket.accessToken,
		}).Do(ctx)

	if err != nil {
		return "", errors.Wrapf(err, "add failed: %s", url)
	}

	var response struct {
		Item struct {
			ItemID string `json:"item_id"`
		} `json:"item"`
	}

	defer resp.Body.Close()
	if err := resp.JSON(&response); err != nil {
		return "", err
	}

	return response.Item.ItemID, nil
}

// Delete delete article by item id
// NOTE Delete action always success ㅡㅡ;
func (a *ArticlesAPI) Delete(ctx context.Context, itemIDs ...string) error {
	log.Debugf("remove item: %s", itemIDs)

	params := make([]actionParam, len(itemIDs))
	for i := 0; i < len(itemIDs); i++ {
		params[i].Action = "delete"
		params[i].ItemID = itemIDs[i]
	}

	if _, err := a.sendAction(ctx, params); err != nil {
		return errors.Wrapf(err, "delete(%s)", itemIDs)
	}

	return nil
}
