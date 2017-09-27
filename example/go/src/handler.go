package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/nfnt/resize"
)

var outputFolder string

func init() {
	outputFolder = os.Getenv("OUTPUT_FOLDER")
	if outputFolder == "" {
		outputFolder = "files"
	}
}

func Handle(evt interface{}, ctx *runtime.Context) (string, error) {
	// fmt.Fprintln(os.Stderr, "GO Handle called:", ctx.AWSRequestID)
	// fmt.Println("Printing to stout:", ctx.AWSRequestID)
	// return "Hello, World!", nil
	evtmap, ok := evt.(map[string]interface{})
	if !ok {
		fmt.Println("Wrong type: (%v) %T", evt, evt)
		return "", fmt.Errorf("wrong type for event %T", evt)
	}

	bucket, key, region, err := getBucketAndKey(evtmap)
	if err != nil {
		return "", err
	}

	fmt.Printf("handling %s:%s\n", bucket, key)
	nkey, err := newName(key)
	if err != nil {
		return "", err
	}

	sess := session.New(&aws.Config{Region: &region})
	svc := s3.New(sess)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return "", fmt.Errorf("failed getting object: %s", err)
	}
	if result.Body == nil {
		return "", errors.New("No file body")
	}
	defer result.Body.Close()

	img, format, err := image.Decode(result.Body)
	if err != nil {
		fmt.Println("Not an image file: %s", err)
		return "moved", move(svc, bucket, key, nkey)
	}
	fmt.Printf("Image %s - %s\n", format, img.Bounds().String())

	if img.Bounds().Dx() <= 300 && img.Bounds().Dy() <= 300 {
		fmt.Println("does not need resizing")
		return "Moved", move(svc, bucket, key, nkey)
	}

	var resized image.Image
	if img.Bounds().Dx() > img.Bounds().Dy() {
		resized = resize.Resize(300, 0, img, resize.Bilinear)
	} else {
		resized = resize.Resize(0, 300, img, resize.Bilinear)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, nil); err != nil {
		return "", err
	}
	fmt.Println("Encoded, len:", buf.Len())

	// We encoded it as jpeg, lets make sure that the name fit
	if ext := filepath.Ext(nkey); ext != "jpeg" && ext != "jpg" {
		nkey = nkey + ".jpg"
	}

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(nkey),
		Body:        &buf,
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", err
	}
	attemptDelete(svc, bucket, key)
	return fmt.Sprintf("Resized %s -> (%d,%d)", key, resized.Bounds().Dx(), resized.Bounds().Dy()), nil
}

func move(svc *s3.S3, bucket, from, to string) error {
	_, err := svc.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(to),
		CopySource: aws.String(bucket + "/" + from),
	})
	if err != nil {
		return err
	}
	attemptDelete(svc, bucket, from)
	return nil
}

func attemptDelete(svc *s3.S3, bucket, key string) {
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println("Error deleteing object: ", err)
	}
}

func newName(key string) (string, error) {
	return path.Join(outputFolder, path.Base(key)), nil
}
