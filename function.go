package bitbucket_webhook_golang

import (
	"bytes"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"fmt"
	"gopkg.in/go-playground/webhooks.v5/bitbucket"
	"io/ioutil"
	"net/http"
	"strings"
)

var webHooksUrl = "https://chat.googleapis.com/v1/spaces/xxxxxx/messages?key=yyyyyyyyyyyy"
var databaseUrl = "https://url-to-realtime-database/"
var database db.Client
var ctx = context.Background()

func init() {
	config := &firebase.Config{DatabaseURL: databaseUrl}
	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		fmt.Println("error initializing app:", err)
	}

	client, err := app.Database(ctx)
	if err != nil {
		fmt.Println("Error initializing database client:", err)
	}
	database = *client
}

func pushToGoogleChat(message string, threadId *string) (string, error) {
	var chatRequest Chat
	if threadId == nil {
		chatRequest = Chat{Text: message, Thread: Thread{Name: ""}}
	} else {
		chatRequest = Chat{Text: message, Thread: Thread{Name: *threadId}}
	}

	jsonReq, err := json.Marshal(chatRequest)
	resp, err := http.Post(webHooksUrl, "application/json; charset=utf-8", bytes.NewBuffer(jsonReq))
	if err != nil {
		fmt.Println(err)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	body := Chat{}

	if err := json.Unmarshal(data, &body); err != nil {
		return "", err
	}
	return body.Thread.Name, nil
}

func savePullRequestId(pullRequestId string, threadId string) {
	err := database.NewRef("chatThread").Child(pullRequestId).Set(ctx, threadId)
	if err != nil {
		fmt.Println(err)
	}
}

func getThreadIdByPullRequestId(pullRequestId string) (string, error) {
	var data string
	err := database.NewRef("chatThread").Child(pullRequestId).Get(ctx, &data)
	if err != nil {
		return "", err
	}
	return data, nil
}

func removeThreadByPullRequestId(pullRequestId string) {
	err := database.NewRef("chatThread").Child(pullRequestId).Delete(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func getPullRequestId(request bitbucket.PullRequest) string {
	return strings.Split(request.Links.HTML.Href, "/bitbucket.org/")[1]
}

func pullRequestCreated(body bitbucket.PullRequestCreatedPayload) {
	message := "<users/all>\n" +
		"Repo :  " + strings.TrimSpace(body.Repository.Name) + "\n" +
		"Title :     " + strings.TrimSpace(body.PullRequest.Title) + "\n" +
		"Branch : " + strings.TrimSpace(body.PullRequest.Source.Branch.Name) + "   >   " + strings.TrimSpace(body.PullRequest.Destination.Branch.Name) + "\n" +
		"Author : _" + strings.TrimSpace(body.Actor.DisplayName) + "_\n" +
		"Link :     <" + body.PullRequest.Links.HTML.Href + "|" + body.PullRequest.Links.HTML.Href + ">\n" +
		"มีออเดอร์ใหม่เข้ามา คุณต้องเร่งมือแล้วนะครับ"
	threadId, err := pushToGoogleChat(message, nil)
	if err != nil {
		fmt.Println(err)
	}
	savePullRequestId(getPullRequestId(body.PullRequest), threadId)
}

func pullRequestUpdated(body bitbucket.PullRequestUpdatedPayload) {
	message := "_" + strings.TrimSpace(body.Actor.DisplayName) + "_ has Updated."
	threadId, err := getThreadIdByPullRequestId(getPullRequestId(body.PullRequest))
	if err != nil {
		fmt.Println("cannot get thread id: ", err)
	} else {
		_, _ = pushToGoogleChat(message, &threadId)
	}
}

func pullRequestCommented(body bitbucket.PullRequestCommentCreatedPayload) {
	message := "_" + strings.TrimSpace(body.Actor.DisplayName) + "_ has Commented.\n" +
		"เชฟเตือนแล้วนะ!"
	threadId, err := getThreadIdByPullRequestId(getPullRequestId(body.PullRequest))
	if err != nil {
		fmt.Println("cannot get thread id: ", err)
	} else {
		_, _ = pushToGoogleChat(message, &threadId)
	}
}

func pullRequestApproved(body bitbucket.PullRequestApprovedPayload) {
	message := "_" + strings.TrimSpace(body.Actor.DisplayName) + "_ has Approved.\n" +
		"ให้ผ่านครับเชฟ"
	threadId, err := getThreadIdByPullRequestId(getPullRequestId(body.PullRequest))
	if err != nil {
		fmt.Println("cannot get thread id: ", err)
	} else {
		_, _ = pushToGoogleChat(message, &threadId)
	}
}

func pullRequestUnapproved(body bitbucket.PullRequestUnapprovedPayload) {
	message := "_" + strings.TrimSpace(body.Actor.DisplayName) + "_ has Unapproved.\n" +
		"คุณต้องตั้งสติและรีบแก้ปัญหาเฉพาะหน้านะครับ"
	threadId, err := getThreadIdByPullRequestId(getPullRequestId(body.PullRequest))
	if err != nil {
		fmt.Println("cannot get thread id: ", err)
	} else {
		_, _ = pushToGoogleChat(message, &threadId)
	}
}

func pullRequestMerged(body bitbucket.PullRequestMergedPayload) {
	message := "_" + strings.TrimSpace(body.Actor.DisplayName) + "_ has Merged.\n" +
		"เสริฟอาหารแล้วครับเชฟ"
	threadId, err := getThreadIdByPullRequestId(getPullRequestId(body.PullRequest))
	if err != nil {
		fmt.Println("cannot get thread id: ", err)
	} else {
		_, _ = pushToGoogleChat(message, &threadId)
	}
	removeThreadByPullRequestId(getPullRequestId(body.PullRequest))
}

func pullRequestDeclined(body bitbucket.PullRequestDeclinedPayload) {
	message := "_" + strings.TrimSpace(body.Actor.DisplayName) + "_ has Declined.\n" +
		"ออเดอร์โดนยกเลิกแล้วครับเชฟ"
	threadId, err := getThreadIdByPullRequestId(getPullRequestId(body.PullRequest))
	if err != nil {
		fmt.Println("cannot get thread id: ", err)
	} else {
		_, _ = pushToGoogleChat(message, &threadId)
	}
	removeThreadByPullRequestId(getPullRequestId(body.PullRequest))
}

func PullRequest(w http.ResponseWriter, r *http.Request) {
	webHook, err := bitbucket.New()
	if err != nil {
		return
	}
	request, err := webHook.Parse(r, bitbucket.Event(r.Header.Get("X-Event-Key")))
	switch requestType := r.Header.Get("X-Event-Key"); requestType {
	case "pullrequest:created":
		pullRequestCreated(request.(bitbucket.PullRequestCreatedPayload))
	case "pullrequest:updated":
		pullRequestUpdated(request.(bitbucket.PullRequestUpdatedPayload))
	case "pullrequest:comment_created":
		pullRequestCommented(request.(bitbucket.PullRequestCommentCreatedPayload))
	case "pullrequest:approved":
		pullRequestApproved(request.(bitbucket.PullRequestApprovedPayload))
	case "pullrequest:unapproved":
		pullRequestUnapproved(request.(bitbucket.PullRequestUnapprovedPayload))
	case "pullrequest:fulfilled":
		pullRequestMerged(request.(bitbucket.PullRequestMergedPayload))
	case "pullrequest:rejected":
		pullRequestDeclined(request.(bitbucket.PullRequestDeclinedPayload))
	}

	_, _ = w.Write([]byte("OK"))
}
