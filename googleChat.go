package bitbucket_webhook_golang

type Chat struct {
	Text   string `json:"text"`
	Thread Thread `json:"thread"`
}

type Thread struct {
	Name string `json:"name"`
}
