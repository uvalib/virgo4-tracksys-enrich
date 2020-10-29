package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Proxy contains methods for accessing the S3 cache
type S3Proxy struct {
	bucketName string
	uploader   *s3manager.Uploader
}

// NewS3Proxy creates a new S3 proxy object
func NewS3Proxy(cfg *ServiceConfig) *S3Proxy {

	proxy := S3Proxy{}
	proxy.bucketName = cfg.DigitalContentCacheBucket
	sess, err := session.NewSession()
	if err == nil {
		proxy.uploader = s3manager.NewUploader(sess)
	}

	return &proxy
}

// WriteToCache writes the contents of the specified cache element
func (s3p *S3Proxy) WriteToCache(key string, content string) error {

	contentSize := len(content)
	destname := fmt.Sprintf("s3://%s/%s", s3p.bucketName, key)
	log.Printf("INFO: uploading to %s (%d bytes)", destname, contentSize)

	upParams := s3manager.UploadInput{
		Bucket: &s3p.bucketName,
		Key:    &key,
		Body:   bytes.NewReader([]byte(content)),
	}

	// Perform an upload.
	start := time.Now()
	_, err := s3p.uploader.Upload(&upParams)
	if err != nil {
		log.Printf("ERROR: uploading to %s (%s)", destname, err.Error())
		return err
	}

	duration := time.Since(start)
	log.Printf("INFO: upload of %s complete in %0.2f seconds", destname, duration.Seconds())

	return nil
}

//
// end of file
//
