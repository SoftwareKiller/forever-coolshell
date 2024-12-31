package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

//go:embed assets
var Assets embed.FS

//go:embed content/haoel
var Author embed.FS

//go:embed content/articles
var Articles embed.FS

//go:embed content/list
var List embed.FS

//go:embed uploads
var Uploads embed.FS

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/page/1.html")
	})

	r.GET("/search", searchHandler)

	assets, _ := fs.Sub(Assets, "assets")
	r.StaticFS("/assets", http.FS(assets))

	author, _ := fs.Sub(Author, "content/haoel")
	r.StaticFS("/haoel", http.FS(author))

	articles, _ := fs.Sub(Articles, "content/articles")
	r.StaticFS("/articles", http.FS(articles))

	page, _ := fs.Sub(List, "content/list")
	r.StaticFS("/page", http.FS(page))

	uploads, _ := fs.Sub(Uploads, "uploads")
	r.StaticFS("/uploads", http.FS(uploads))

	log.Println("[酷壳 Cool Shell Forever 电子存档]")
	log.Println()
	log.Println("芝兰生于深谷，不以无人而不芳")
	log.Println("君子修身养德，不以穷困而改志")
	log.Println()
	log.Println("感恩皓叔的无私分享，以及总在最需要的时刻愿意花时间给予指导和帮助。")
	log.Println()
	log.Println("工具使用：")
	log.Println("    如果你需要改变端口，可以使用环境变量 PORT 来指定端口，例如：")
	log.Println("    PORT=8080 ./forever-coolshell ")

	port := "8080"
	portEnv := os.Getenv("PORT")
	p, err := strconv.ParseInt(portEnv, 10, 64)
	if err != nil {
		log.Println("使用默认端口", port)
	} else {
		port = fmt.Sprintf("%d", p)
		log.Println("使用指定端口", port)
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func searchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	results := searchArticles(query)
	c.HTML(http.StatusOK, "search_results.html", gin.H{
		"query":   query,
		"results": results,
	})
}

type SearchResult struct {
	Link        string    `json:"link"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func searchArticles(query string) []SearchResult {
	var results []SearchResult

	fs.WalkDir(Articles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, err := fs.ReadFile(Articles, path)
			if err != nil {
				return err
			}
			if strings.Contains(string(content), query) {
				link := strings.Replace(path, "content", "", 1)
				description := getScreenReaderText(string(content))
				date := extractDate(string(content))
				results = append(results, SearchResult{Link: link, Description: description, Date: date})
			}
		}
		return nil
	})

	// Sort results by date in descending order
	sort.Slice(results, func(i, j int) bool {
		return results[i].Date.After(results[j].Date)
	})

	return results
}

func getScreenReaderText(content string) string {
	// Extract the content of screen-reader-text
	start := strings.Index(content, "screen-reader-text")
	if start == -1 {
		return ""
	}
	start = strings.Index(content[start:], ">") + start + 1
	end := strings.Index(content[start:], "<") + start
	if start == -1 || end == -1 {
		return ""
	}
	return content[start:end]
}

func extractDate(content string) time.Time {
	// Extract the datetime attribute from the content
	start := strings.Index(content, "datetime=\"")
	if start == -1 {
		return time.Time{}
	}
	start += len("datetime=\"")
	end := strings.Index(content[start:], "\"")
	if end == -1 {
		return time.Time{}
	}
	dateStr := content[start : start+end]
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}
	}
	return date
}
