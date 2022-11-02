package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Payload struct {
	Body string `json:"body"`
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func LookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func main() {
	var giteaToken string
	var giteaAddress string
	var comment string
	var commentFile string
	var commentIsCode bool
	var repoOwner string
	var repoName string
	var prIndex int

	flag.StringVar(&giteaToken, "gitea-token", LookupEnvOrString("PLUGIN_GITEA_TOKEN", giteaToken), "API token for Gitea")
	flag.StringVar(&giteaAddress, "gitea-address", LookupEnvOrString("PLUGIN_GITEA_ADDRESS", giteaAddress), "Gitea URL")
	flag.StringVar(&comment, "comment", LookupEnvOrString("PLUGIN_COMMENT", comment), "Comment for Gitea")
	flag.StringVar(&commentFile, "comment-file", LookupEnvOrString("PLUGIN_COMMENT_FILE", commentFile), "Use file as comment for Gitea")
	flag.BoolVar(&commentIsCode, "commentIsCode", LookupEnvOrBool("PLUGIN_COMMENT_IS_CODE", commentIsCode), "Wrap the comment in a code block")
	flag.StringVar(&repoOwner, "repo-owner", LookupEnvOrString("CI_REPO_OWNER", repoOwner), "Owner of the repository")
	flag.StringVar(&repoName, "repo-name", LookupEnvOrString("CI_REPO_NAME", repoName), "Name of the repository")
	flag.IntVar(&prIndex, "pr-index", LookupEnvOrInt("CI_COMMIT_PULL_REQUEST", prIndex), "Index of the PR")

	flag.Parse()

	if comment != "" && commentFile != "" {
		fmt.Println(comment, commentFile)
		panic("Cannot specify both comment and comment file")
	}

	if comment == "" && commentFile == "" {
		panic("You must provide a comment or comment file")
	}

	if giteaToken == "" {
		panic("You must provide a Gitea API Token")
	}
	if giteaAddress == "" {
		panic("You must provide a Gitea URL")
	}
	if repoOwner == "" {
		panic("You must provide an repo owner")
	}
	if repoName == "" {
		panic("You must provide a repo name")
	}
	if prIndex == 0 {
		panic("You must provide an index for PR")
	}

	if comment != "" && commentIsCode {
		comment = fmt.Sprintf("```\n%s\n```", comment)
	}

	if commentFile != "" {
		file, err := os.Open(commentFile)
		if err != nil {
			panic(err)
		}
		commentBytes, _ := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		comment = string(commentBytes)
		if commentIsCode {
			comment = fmt.Sprintf("```\n%s\n```", comment)
		}
	}

	data := Payload{
		Body: comment,
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	body := bytes.NewReader(payloadBytes)

	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/issues/%d/comments?access_token=%s", giteaAddress, repoOwner, repoName, prIndex, giteaToken)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
