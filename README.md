# Bitbucket webhook Google Chat Golang
This repository is an example how to integrate Bitbucket with Google Chat 
using [Firebase](https://firebase.google.com) functions.

1. Install Google Cloud SDK
    https://cloud.google.com/sdk/docs/install
    
2. Login gcloud
    ```$bash
    gcloud init
    ```
3. Download dependency
    ```$bash
    go mod vendor
    ```
4. Deploy to firebase
    ```$bash
    gcloud functions deploy pullrequest --runtime go111 --project <project_id> --trigger-http
    ```

## Example message when create pull request
```
@all
Title  :    Merge title message
Branch :    develop   >   master
Author :    Shimada Genji
Link   :    https://bitbucket.org/link/to/pull-requests/2
```