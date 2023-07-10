package s3

import (
	"file-management-service/config"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3 represents the Amazon S3 service.
type S3 struct {
	bucketName string
	svc        *s3.S3
}

type ListFilesResponseType struct {
	Files         []ObjectDetails
	NextPageToken string
}

// NewS3 creates a new S3 instance with the specified bucket name and AWS session.
func NewClient(config *config.Config) (*S3, error) {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region), // Replace with your desired AWS region,
		Credentials: credentials.NewStaticCredentials(
			config.AwsAccessKeyID,     // Replace with your AWS access key ID
			config.AwsSecretAccessKey, // Replace with your AWS secret access key
			"",
		),
	})

	if err != nil {
		return nil, err
	}

	// Create an S3 service client
	svc := s3.New(sess)

	return &S3{
		bucketName: config.BucketName,
		svc:        svc,
	}, nil
}

// UploadFile uploads a file to the S3 bucket.
// UploadFile uploads a file to the S3 bucket.
func (s *S3) UploadFile(src io.Reader, objectKey string) error {
	// Upload the file to S3
	_, err := s.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
		Body:   aws.ReadSeekCloser(src),
	})
	if err != nil {
		return err
	}

	fmt.Println("File uploaded successfully")
	return nil
}

// ListObjects lists all the objects within a folder in the S3 bucket.
func (s *S3) ListFiles(folderPath string, nextPageToken string, pageSize int) (*ListFilesResponseType, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucketName),
		Prefix:  aws.String(folderPath),
		MaxKeys: aws.Int64(int64(pageSize)),
	}

	if nextPageToken != "" {
		input.ContinuationToken = aws.String(nextPageToken)
	}

	resp, err := s.svc.ListObjectsV2(input)
	if err != nil {
		return nil, err
	}

	// send all file details
	var objects []ObjectDetails
	for _, obj := range resp.Contents {
		objects = append(objects, ObjectDetails{
			Name:         *obj.Key,
			IsFolder:     *obj.Size == 0,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
		})
	}

	nextToken := ""
	if resp.NextContinuationToken != nil {
		nextToken = *resp.NextContinuationToken
	}

	response := &ListFilesResponseType{
		Files:         objects,
		NextPageToken: nextToken,
	}

	return response, nil
}

func (s *S3) ListFolders(folderPath string, nextPageToken string, pageSize int) (*ListFilesResponseType, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucketName),
		Prefix:  aws.String(folderPath),
		MaxKeys: aws.Int64(int64(pageSize)),
	}

	if nextPageToken != "" {
		input.ContinuationToken = aws.String(nextPageToken)
	}

	resp, err := s.svc.ListObjectsV2(input)
	if err != nil {
		return nil, err
	}

	// send all file details
	var objects []ObjectDetails
	for _, obj := range resp.Contents {
		// filter out the folders (size == 0) means folder
		if *obj.Size == 0 {
			objects = append(objects, ObjectDetails{
				Name:         *obj.Key,
				IsFolder:     *obj.Size == 0,
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
			})
		}
	}

	nextToken := ""
	if resp.NextContinuationToken != nil {
		nextToken = *resp.NextContinuationToken
	}

	response := &ListFilesResponseType{
		Files:         objects,
		NextPageToken: nextToken,
	}

	return response, nil
}

// GetFile retrieves a file from the specified bucket and key in S3.
func (s *S3) GetFile(bucket, key string) (io.Reader, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s.svc.GetObject(input)
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

// Function to generate a signed download URL for the object
func (s *S3) GenerateDownloadLink(objectKey string) (string, error) {
	req, _ := s.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	})

	downloadURL, err := req.Presign(15 * time.Minute) // Set the validity period of the signed URL
	if err != nil {
		return "", err
	}

	return downloadURL, nil
}

// DeleteObject deletes an object from the S3 bucket.
func (s *S3) DeleteObject(objectKey string) error {
	_, err := s.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return err
	}

	fmt.Println("Object deleted successfully")
	return nil
}

// DeleteFolder deletes a folder and its contents from the S3 bucket.
func (s *S3) DeleteFolder(folderPath string) error {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	response, err := s.ListFiles(folderPath, "", 1000)
	iterator := 0

	for response.NextPageToken != "" {
		iterator++

		for _, obj := range response.Files {
			if obj.IsFolder {
				err = s.DeleteFolder(obj.Name)
				if err != nil {
					return err
				}
			} else {
				err = s.DeleteObject(obj.Name)
				if err != nil {
					return err
				}
			}
		}

		response, err = s.ListFiles(folderPath, response.NextPageToken, 1000)
		if err != nil {
			return err
		}
	}

	fmt.Println("Folder deleted successfully")
	return nil
}

// ListAllFilesAndFolders lists all files and folders within the specified folder path.
func (s *S3) ListAllFilesAndFolders(folderPath string, nextPageToken string, pageSize int) ([]ObjectDetails, error) {
	// re-use the ListFiles function to get the list of files and folders
	response, err := s.ListFiles(folderPath, "", 1000)

	if err != nil {
		return nil, err
	}

	// filter out the folders
	filterResult := []ObjectDetails{}

	for _, obj := range response.Files {
		if obj.IsFolder {
			filterResult = append(filterResult, obj)
		}
	}

	return filterResult, nil
}
