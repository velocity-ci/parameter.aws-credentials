package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

const TaskMetaDataBase = "http://169.254.170.2"

type taskCredentialResp struct {
	RoleARN         string    `json:"RoleArn"`
	AccessKeyID     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	Token           string    `json:"Token"`
	Expiration      time.Time `json:"Expiration"`
}

func getCredentialsFromSts() (export *exported, expires time.Time, _ error) {
	svc := sts.New(session.New())
	input := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(7200),
	}

	result, err := svc.GetSessionToken(input)
	if err != nil {
		return nil, expires, err
	}

	return &exported{
		AccessKeyID:     *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		Token:           *result.Credentials.SessionToken,
	}, *result.Credentials.Expiration, nil
}

func getCredentialsFromTask(httpClient *http.Client) (export *exported, expires time.Time, _ error) {
	relativeURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	if len(relativeURI) < 1 {
		return nil, expires, errors.New("missing AWS_CONTAINER_CREDENTIALS_RELATIVE_URI environment variable")
	}

	address := fmt.Sprintf("%s/%s", TaskMetaDataBase, relativeURI)

	r, err := httpClient.Get(address)
	if err != nil {
		return nil, expires, err
	}

	var creds taskCredentialResp
	json.NewDecoder(r.Body).Decode(&creds)
	r.Body.Close()

	if len(creds.AccessKeyID) < 1 {
		return nil, expires, errors.New("missing credentials in response")
	}

	return &exported{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		Token:           creds.Token,
	}, creds.Expiration, nil
}

const InstanceMetaDataBase = "http://169.254.169.254/latest/meta-data/iam/security-credentials"

type instanceCredentialResp struct {
	Code            string    `json:"Code"`
	LastUpdated     time.Time `json:"LastUpdated"`
	Type            string    `json:"Type"`
	AccessKeyID     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	Token           string    `json:"Token"`
	Expiration      time.Time `json:"Expiration"`
}

func getCredentialsFromInstance(httpClient *http.Client, roleName string) (export *exported, expires time.Time, _ error) {
	if len(roleName) < 1 {
		return nil, expires, errors.New("missing roleName")
	}

	address := fmt.Sprintf("%s/%s", InstanceMetaDataBase, roleName)

	r, err := httpClient.Get(address)
	if err != nil {
		return nil, expires, err
	}

	var creds instanceCredentialResp
	json.NewDecoder(r.Body).Decode(&creds)
	r.Body.Close()

	if len(creds.AccessKeyID) < 1 {
		return nil, expires, errors.New("missing credentials in response")
	}

	return &exported{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		Token:           creds.Token,
	}, creds.Expiration, nil
}

func main() {
	roleName := flag.String("roleName", "", "AWS IAM role name to get credentials for")
	flag.Parse()

	httpClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	out := output{
		Secret: true,
	}
	if creds, expires, err := getCredentialsFromSts(); err == nil {
		out.Exports = creds
		out.Expires = expires
		out.State = "success"
	} else if creds, expires, err := getCredentialsFromTask(httpClient); err == nil {
		out.Exports = creds
		out.Expires = expires
		out.State = "success"
	} else if creds, expires, err := getCredentialsFromInstance(httpClient, *roleName); err == nil {
		out.Exports = creds
		out.Expires = expires
		out.State = "success"
	} else {
		out.State = "critical"
		out.Error = "no valid response from metadata endpoints"
		out.Exports = &exported{}
	}
	o, _ := json.Marshal(out)

	fmt.Printf("%s", o)

}

type output struct {
	Secret  bool      `json:"secret"`
	Exports *exported `json:"exports"`
	Expires time.Time `json:"expires"`
	Error   string    `json:"error"`
	State   string    `json:"state"`
}

type exported struct {
	AccessKeyID     string `json:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `json:"AWS_SECRET_ACCESS_KEY"`
	Token           string `json:"AWS_SESSION_TOKEN"`
}
