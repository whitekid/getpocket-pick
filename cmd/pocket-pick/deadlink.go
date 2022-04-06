package main

import (
	"context"

	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/whitekid/goxp"
	"github.com/whitekid/goxp/fx"
	"github.com/whitekid/goxp/request"
	"github.com/whitekid/pocket-pick/config"
	"github.com/whitekid/pocket-pick/pkg/pocket"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:  "check-dead-link",
		Long: "check dead link",
		RunE: func(cmd *cobra.Command, args []string) error { return checkDeadLink(rootCmd.Context()) },
	})
}

func checkDeadLink(ctx context.Context) error {
	api := pocket.New(config.ConsumerKey(), config.AccessToken())
	items, err := api.Articles.Get().Favorite(pocket.Favorited).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "articles.Get(Favorite)")
	}
	log.Debug("items: %d", len(items))

	ch := make(chan *pocket.Article)

	go func() {
		close(ch)
		notFoundItems := []string{"274841724", "758026316", "392120428", "494194220"}

		fx.ForEachMap(items, func(k string, v *pocket.Article) {
			if fx.Contains(notFoundItems, v.ResolvedID) {
				return
			}

			ch <- v
		})
	}()

	// start 4 worker
	var itemsToDelete []string
	goxp.DoWithWorker(4, func(i int) {
		for article := range ch {
			log.Infof("checking %s %s", article.ItemID, article.ResolvedURL)
			resp, err := request.Get(article.ResolvedURL).
				Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36").
				Do(ctx)
			if err != nil {
				log.Errorf("check link failed: itemID: %s,   link: %s, err: %s", article.ItemID, article.ResolvedURL, err)
				itemsToDelete = append(itemsToDelete, article.ItemID)
				continue
			}

			if !resp.Success() {
				log.Errorf("failed with %d, itemID: %s, link: %s", resp.StatusCode, article.ItemID, article.ResolvedURL)
				itemsToDelete = append(itemsToDelete, article.ItemID)
			}
		}
	})

	if len(itemsToDelete) > 0 {
		log.Infof("deleting: %v", itemsToDelete)

		if err := api.Articles.Delete(ctx, itemsToDelete...); err != nil {
			return errors.Wrapf(err, "articles.Delete(%s)", itemsToDelete)
		}
	}

	return nil
}
