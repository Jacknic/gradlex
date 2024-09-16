package cmd

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const GRADLE_VERSION_URL = "https://services.gradle.org/versions/all"

var listType string

func init() {
	listCmd.Aliases = []string{"ls"}
	listCmd.Flags().StringVarP(&listType, "type", "t", "", "Gradle 版本类型")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "show gradle version list",
	Long:  `show gradle version list`,
	Run: func(cmd *cobra.Command, args []string) {
		getVersionList(listType)
	},
}

// 获取 Gradle 版本列表
func getVersionList(listType string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", GRADLE_VERSION_URL, nil)
	req.Proto = "HTTP/1.1"
	req.Host = "services.gradle.org"
	req.Header = map[string][]string{
		"Accept-Encoding": {"gzip"},
		"User-Agent":      {"curl/7.68.0"},
		"Accept":          {"*/*"},
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("failed to fetch Gradle versions, status code: %d\n", resp.StatusCode)
		return
	}

	// fmt.Printf("fetching Gradle versions... %v\n", resp.ContentLength)
	lastModified := resp.Header.Get("last-modified")
	os.MkdirAll(getGradleUserHome(), os.ModePerm)
	cacheFile := getGradleUserHome() + "/version-all.json"
	info, err := os.Stat(cacheFile)
	if os.IsNotExist(err) {
		f, err := os.Create(cacheFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		info, _ = f.Stat()
	}
	modTime, _ := time.Parse(time.RFC1123, lastModified)

	sameSize := info != nil && resp.ContentLength == info.Size() && resp.ContentLength != 0
	// fmt.Printf("lastModified: %s ContentLength: %v => %v \n ", lastModified, resp.ContentLength, info.Size())
	if info != nil && modTime.Before(info.ModTime()) && sameSize {
		// fmt.Println("use cache file")
	} else {
		// fmt.Println("fetch new version list")
		f, err := os.Create(cacheFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		reader, _ := gzip.NewReader(resp.Body)
		io.Copy(f, reader)
	}
	f, _ := os.Open(cacheFile)
	defer f.Close()
	body, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var versionList []map[string]interface{}
	err = json.Unmarshal(body, &versionList)
	if err != nil {
		fmt.Println("JSON Parse error:", err)
		return
	}

	count := 0
	major := ""
	sort.Slice(versionList, func(i, j int) bool {
		return versionList[i]["version"].(string) < versionList[j]["version"].(string)
	})
	for _, version := range versionList {
		versionName := version["version"].(string)
		if listType != "all" && strings.Contains(versionName, "-") {
			continue
		}
		newMajor := versionName[:1]
		if major != newMajor {
			major = newMajor
			if count != 0 {
				fmt.Print("\n")
			}
			// fmt.Printf("=========  %s  =========\n", major)
		}
		fmt.Printf("%s   ", version["version"])
		count++
	}
	fmt.Printf("\n\ntatal: %d\n", count)
}
