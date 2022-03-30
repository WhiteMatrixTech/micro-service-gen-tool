package pkg

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

var s3Session *s3.S3 = nil

func getS3Session() *s3.S3 {
	// this is hardcoded
	if s3Session == nil {
		s3Session = s3.New(session.Must(session.NewSession(&aws.Config{
			Region: aws.String("us-west-2")},
		)))
	}
	return s3Session
}

func ReadTokenFromS3() (string, error) {

	rawObject, err := getS3Session().GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String("whitematrix-internal"),
			Key:    aws.String("github-access-tokens/code-gen-tool-token.txt"),
		})

	if err != nil {
		log.Println(err)
		return "", err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(rawObject.Body)

	if err != nil {
		log.Println(err)
		return "", err
	}

	fileContentAsString := buf.String()
	return fileContentAsString, nil
}
