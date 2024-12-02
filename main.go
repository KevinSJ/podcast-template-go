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
	"flag"
	"html"
	"io/fs"
	"log"
	"net/url"
	"os"
	"strings"
	"text/template"
)

const (
	TTS_FILE_BIT_RATE  float64 = 32000.0
	DOMAIN_NAME        string  = "https://lab.jiangsc.me/"
	FEED_PATH          string  = "feed.xml"
	POD_TITLE          string  = "My Daily Readings"
	POD_DESC           string  = "Podcast for daily reading"
	DATE_FORMAT        string  = "Mon, 02 Jan 2006 15:04:05 -0700"
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

func getDomainName() string {
	if domain, exist := os.LookupEnv("DOMAIN_NAME"); exist {
		return domain
	}

	return DOMAIN_NAME
}

func main() {
	domainName := flag.String("d", getDomainName(), "usage")
	feedPath := flag.String("f", FEED_PATH, "usage")
	podTitle := flag.String("title", POD_TITLE, "usage")
	podDescription := flag.String("desc", POD_DESC, "usage")
	flag.Parse()

	podcast := Podcast{
		PodLink:        *domainName + *feedPath,
		PodTitle:       *podTitle,
		PodDescription: *podDescription,
	}
	// Prepare some data to insert into the template.
	episodes := []Episode{}

	fs.WalkDir(os.DirFS("."), ".", func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(path, "mp3") {
			fileUrl := *domainName + html.EscapeString(url.PathEscape(path))
			embeddedTitle, artist, _ := ReadID3Tags(path)
			title := ""
			if embeddedTitle != "" {
				title = "[" + artist + "]" + embeddedTitle
			} else {
				title = html.EscapeString(strings.ReplaceAll(de.Name(), ".mp3", ""))
			}

			fileInfo, _ := de.Info()
			pubDate := fileInfo.ModTime().UTC().Format(DATE_FORMAT)
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

	f, err := os.Create("./" + *feedPath)
	if err != nil {
		log.Panic(err)
	}

	t.Execute(f, podcast)
}
