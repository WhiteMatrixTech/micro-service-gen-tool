package pkg

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

func ReadTokenFromS3() (string, error) {
	// this is hardcoded
	s3Session := s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)))

	rawObject, err := s3Session.GetObject(
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
