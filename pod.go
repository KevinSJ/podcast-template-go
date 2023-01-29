/*
 *podcast template go
 *This software will create podcast xml based on a template
 *Copyright © 2023 KevinSJ
 *
 *Permission is hereby granted, free of charge, to any person obtaining
 *a copy of this software and associated documentation files (the "Software"),
 *to deal in the Software without restriction, including without limitation
 *the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *and/or sell copies of the Software, and to permit persons to whom the
 *Software is furnished to do so, subject to the following conditions:
 *
 *The above copyright notice and this permission notice shall be included
 *in all copies or substantial portions of the Software.
 *
 *THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 *EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
 *OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 *DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 *TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
 *OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"html"
	"io/fs"
	"log"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	TTS_FILE_BIT_RATE  float64 = 32000.0
	DOMAIN_NAME        string  = "https://lab.jiangsc.me/"
	DATE_FORMAT        string  = "Mon Jan 2 15:04:05 MST 2006"
	BIT_TO_BYTE_FACTOR float64 = 8.0
)

type Episode struct {
	Url, Title, Description, PubDate string
	FileSize                         int64
	Duration                         float64
}

type Podcast struct {
	PodLink, PodTitle, PodDescription string
	PodEpisodes                       []Episode
}

func main() {
	DOMAIN_NAME := func() string {
		if domain, exist := os.LookupEnv("DOMAIN_NAME"); exist {
			return domain
		}
		return DOMAIN_NAME
	}()

	currentTime := time.Now().Format(DATE_FORMAT)

	var podcast = Podcast{
		PodLink:        DOMAIN_NAME + "feed.xml",
		PodTitle:       "My Daily Readings",
		PodDescription: "Podcasts for daily",
	}
	// Prepare some data to insert into the template.
	var episodes = []Episode{}

	fs.WalkDir(os.DirFS("."), ".", func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(path, "mp3") {
			fileUrl := DOMAIN_NAME + html.EscapeString(url.PathEscape(path))
			title := html.EscapeString(strings.ReplaceAll(de.Name(), ".mp3", ""))
			pubDate := currentTime
			fileInfo, _ := de.Info()
			fileSize := fileInfo.Size()
			episodeDuration := float64(fileSize) * BIT_TO_BYTE_FACTOR / TTS_FILE_BIT_RATE

			episodes = append(episodes, Episode{
				Url:         fileUrl,
				Title:       title,
				Description: title,
				PubDate:     pubDate,
				FileSize:    fileSize,
				Duration:    episodeDuration,
			})
		}
		return nil
	})

	podcast.PodEpisodes = episodes
	// Create a new template and parse the letter into it.
	t, err := template.ParseFiles("./feed.template.rss")
	if err != nil {
		log.Panic(err)
	}

	f, err := os.Create("./feed.xml")

	if err != nil {
		log.Panic(err)
	}

	t.Execute(f, podcast)
}
