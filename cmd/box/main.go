package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/3uxi/steam-box/pkg/steambox"
	"github.com/google/go-github/github"
)

func main() {
	steamAPIKey := os.Getenv("STEAM_API_KEY")
	steamID, _ := strconv.ParseUint(os.Getenv("STEAM_ID"), 10, 64)
	appIDs := os.Getenv("APP_ID")
	appIDList := make([]uint32, 0)

	for _, appID := range strings.Split(appIDs, ",") {
		appid, err := strconv.ParseUint(appID, 10, 32)
		if err != nil {
			continue
		}
		appIDList = append(appIDList, uint32(appid))
	}

	ghToken := os.Getenv("GH_TOKEN")
	ghUsername := os.Getenv("GH_USER")
	allTimeGistID := os.Getenv("ALL_TIME_GIST_ID")
	recentTimeGistID := os.Getenv("RECENT_TIME_GIST_ID")

	updateOption := os.Getenv("UPDATE_OPTION") // options for update: GIST,MARKDOWN,GIST_AND_MARKDOWN
	markdownFile := os.Getenv("MARKDOWN_FILE") // the markdown filename

	var updateGist, updateMarkdown bool
	if updateOption == "MARKDOWN" {
		updateMarkdown = true
	} else if updateOption == "GIST_AND_MARKDOWN" {
		updateGist = true
		updateMarkdown = true
	} else {
		updateGist = true
	}

	box := steambox.NewBox(steamAPIKey, ghUsername, ghToken)

	ctx := context.Background()

	allTimeLines, err := box.GetPlayTime(ctx, steamID, appIDList...)
	if err != nil {
		panic("GetPlayTime err:" + err.Error())
	}

	recentTimeLines, err := box.GetRecentPlayGanesWithTime(ctx, steamID, 5)
	if err != nil {
		panic("GetRecentTime err:" + err.Error())
	}

	type info struct {
		gistID   string
		lines    []string
		filename string
	}

	tasks := []info{info{allTimeGistID, allTimeLines, "🎮 Steam playtime leaderboard"},
		info{recentTimeGistID, recentTimeLines, "🎮 Steam recent games leaderboard"}}

	for _, v := range tasks {

		if updateGist {
			gist, err := box.GetGist(ctx, v.gistID)
			if err != nil {
				panic("GetGist err:" + err.Error())
			}

			f := gist.Files[github.GistFilename(v.filename)]

			f.Content = github.String(strings.Join(v.lines, "\n"))
			gist.Files[github.GistFilename(v.filename)] = f

			err = box.UpdateGist(ctx, v.gistID, gist)
			if err != nil {
				panic("UpdateGist err:" + err.Error())
			}
		}

		if updateMarkdown && markdownFile != "" {
			title := v.filename
			if updateGist {
				title = fmt.Sprintf(`#### <a href="https://gist.github.com/%s" target="_blank">%s</a>`, v.gistID, title)
			}

			content := bytes.NewBuffer(nil)
			content.WriteString(strings.Join(v.lines, "\n"))

			err = box.UpdateMarkdown(ctx, title, markdownFile, content.Bytes())
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("updating markdown successfully on", markdownFile)
		}
	}
}
