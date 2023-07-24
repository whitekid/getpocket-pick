package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/whitekid/getpocket"
	"github.com/whitekid/goxp/fx"
	"github.com/whitekid/goxp/log"

	"pocket-pick/config"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:          "delete item_id_or_url",
		Long:         "delete article",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return deleteArticle(cmd.Context(), args...) },
	})
}

func deleteArticle(ctx context.Context, idOrURLs ...string) error {
	api := getpocket.New(config.ConsumerKey(), config.AccessToken())
	for _, idOrURL := range idOrURLs {
		// delete by url
		if strings.HasPrefix(idOrURL, "http://") || strings.HasPrefix(idOrURL, "https://") {
			items, err := api.Articles().Get().Search(idOrURL).Do(ctx)
			if err != nil {
				return errors.Wrapf(err, "articles.Get(%+v)", idOrURL)
			}

			if len(items) != 1 {
				return fmt.Errorf("not found: %s", idOrURL)
			}

			ids := fx.Map(fx.Values(items), func(e *getpocket.Article) string { return e.ItemID })
			log.Infof("deleting %s", ids)
			if _, err := api.Modify().Delete(ids...).Do(ctx); err != nil {
				return errors.Wrapf(err, "articles.Delete(%s)", ids)
			}
		} else {
			// delete by id
			if _, err := strconv.Atoi(idOrURL); err != nil {
				return fmt.Errorf("%s is not valid id", idOrURL)
			}

			log.Infof("deleting item %s", idOrURL)
			if _, err := api.Modify().Delete(idOrURL).Do(ctx); err != nil {
				return errors.Wrapf(err, "articles.Delete(%s)", idOrURL)
			}
		}
	}

	return nil
}
