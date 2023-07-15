package routes

import (
	"errors"
	"file-management-service/config"
	"file-management-service/pkg/s3"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all the routes for the application
func RegisterRoutes(e *echo.Echo, config *config.Config) {
	// Define route for uploading images
	e.POST("/upload", func(c echo.Context) error {
		return uploadFileHandler(c, config)
	})

	// Define route for serving files
	e.GET("/download", func(c echo.Context) error {
		return downloadFileHandler(c, config)
	})

	// Delete File
	e.DELETE("/delete", func(c echo.Context) error {
		return deleteFileHandler(c, config)
	})

	e.DELETE("/deleteFolder", func(c echo.Context) error {
		return deleteFolderHandler(c, config)
	})

	// List files and folders within current folder
	e.GET("/list", func(c echo.Context) error {
		return listFolderContentHandler(c, config)
	})

	// List files within current folder -> (nesting not supported)
	e.GET("/files", func(c echo.Context) error {
		return listFilesHandler(c, config)
	})

	// List folders within current folder -> (nesting not supported)
	e.GET("/folders", func(c echo.Context) error {
		return listFoldersHandler(c, config)
	})

	// Define route for testing the server
	e.GET("/ping", ping)
}

// Handler for image upload
func uploadFileHandler(c echo.Context, config *config.Config) error {
	// Get the bucket and folderPath from the form fields
	bucket := c.FormValue("bucket")
	folderPath := c.FormValue("folderPath")

	// Validate the required fields
	if bucket == "" {
		response := s3.GetFailureResponse(errors.New("\"bucket\" is required"))
		return c.JSON(http.StatusInternalServerError, response)
	}
	file, err := c.FormFile("file")
	if err != nil {
		// Handle the error and return an error response
		errorMessage := fmt.Sprintf("Failed to retrieve uploaded file: %s", err.Error())
		response := s3.GetFailureResponse(errors.New(errorMessage))
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		// Handle the error and return an error response
		errorMessage := fmt.Sprintf("Failed to open uploaded file: %s", err.Error())
		response := s3.GetFailureResponse(errors.New(errorMessage))
		return c.JSON(http.StatusInternalServerError, response)
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil {
			// Handle the error (optional)
			fmt.Println("Failed to close uploaded file:", closeErr)
		}
	}()

	// Create a new S3 client
	client, err := s3.NewClient(config)
	if err != nil {
		// Handle the error and return an error response
		errorMessage := fmt.Sprintf("Failed to create S3 client: %s", err.Error())
		response := s3.GetFailureResponse(errors.New(errorMessage))
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Use the file name as it is as the object key
	objectKey := file.Filename
	// Add the folder details
	if folderPath != "" {
		if string(folderPath[len(folderPath)-1]) == "/" {
			objectKey = folderPath + objectKey
		} else {
			objectKey = folderPath + "/" + objectKey
		}
	}

	// Upload the file to S3
	err = client.UploadFile(src, objectKey)
	if err != nil {
		// Handle the error and return an error response
		errorMessage := fmt.Sprintf("Failed to upload file to S3: %s", err.Error())
		response := s3.GetFailureResponse(errors.New(errorMessage))
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Return a success response
	successMessage := fmt.Sprintf("File uploaded successfully with object key: %s", objectKey)
	response := s3.GetUploadDeleteSuccessResponse(successMessage)
	// Return the array of file and folder information as JSON response
	return c.JSON(http.StatusOK, response)
}

// Handler for listing files and folders with pagination [limit & page]
func listFolderContentHandler(c echo.Context, config *config.Config) error {
	// Retrieve the nested folder path from the request parameter
	nestedFolderPath := c.QueryParam("folder")

	// If no folder is provided, set the nested folder path to an empty string
	if nestedFolderPath == "*" {
		nestedFolderPath = ""
	}

	// Create a new S3 client
	client, err := s3.NewClient(config) // Update with your desired region

	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	// List all the files and folders within the nested folder
	objects, err := client.ListAllFilesAndFolders(nestedFolderPath, "", 1000)

	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusOK, objects)
}

// List all files from a folder path
func listFilesHandler(c echo.Context, config *config.Config) error {
	// Retrieve the nested folder path from the request parameter
	nestedFolderPath := c.QueryParam("path")

	// Next page token for pagination
	nextPageToken := c.Request().Header.Get("x-next")

	// Page size for pagination
	pageSize, err := strconv.Atoi(c.QueryParam("pageSize"))
	if err != nil {
		pageSize = config.PaginationPageSize
	}

	// If no folder is provided, set the nested folder path to an empty string
	if nestedFolderPath == "*" {
		nestedFolderPath = ""
	}

	// Create a new S3 client
	client, err := s3.NewClient(config) // Update with your desired region

	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	// List all the files and folders within the nested folder
	objects, err := client.ListFiles(nestedFolderPath, nextPageToken, pageSize)

	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusOK,
		s3.ListFilesResponse{
			Data:                s3.SortFiles(*objects.Data, c),
			NextPageToken:       objects.NextPageToken,
			IsLastPage:          objects.IsLastPage,
			NoOfRecordsReturned: objects.NoOfRecordsReturned,
		},
	)
}

// List all folders from a folder path
func listFoldersHandler(c echo.Context, config *config.Config) error {
	// Retrieve the nested folder path from the request parameter
	nestedFolderPath := c.QueryParam("path")

	// pageSize := config.PaginationPageSize
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	if pageSize == 0 {
		pageSize = config.PaginationPageSize
	}

	// If no folder is provided, set the nested folder path to an empty string
	if nestedFolderPath == "*" {
		nestedFolderPath = ""
	}

	// Create a new S3 client
	client, err := s3.NewClient(config) // Update with your desired region

	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	// List all the files and folders within the nested folder
	objects, err := client.ListFolders(nestedFolderPath, "", pageSize)

	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusOK, objects)
}

// Handler for downloading a file
func downloadFileHandler(c echo.Context, config *config.Config) error {
	bucket := c.QueryParam("bucket") // Correct parameter name: "bucket"
	key := c.QueryParam("objectKey")

	// Create a new S3 client
	client, err := s3.NewClient(config) // Update with your desired region
	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Retrieve the file from S3 using the GetFile function
	file, err := client.GetFile(bucket, key)
	if err != nil {
		// Handle the error
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}
	defer func() {
		if closer, ok := file.(io.Closer); ok {
			closer.Close()
		}
	}()

	// Get the fileName, ignoring folders in prefix.
	fileName := filepath.Base(key)

	// Set the response headers
	c.Response().Header().Set("Content-Type", "application/octet-stream")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	// Copy the file content to the response body
	_, err = io.Copy(c.Response().Writer, file)
	if err != nil {
		// Handle the error
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	return nil
}

func deleteFileHandler(c echo.Context, config *config.Config) error {
	// bucket := c.QueryParam("bucket")
	objectKey := c.QueryParam("objectKey")

	// Create a new S3 client
	client, err := s3.NewClient(config) // Update with your desired region
	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Delete the file or folder from the S3 bucket
	err = client.DeleteObject(objectKey)
	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Return a success response
	response := s3.GetUploadDeleteSuccessResponse("File deleted successfully")
	return c.JSON(http.StatusOK, response)
}

func deleteFolderHandler(c echo.Context, config *config.Config) error {
	// bucket := c.QueryParam("bucket")
	folderPath := c.QueryParam("folderPath")

	// Create a new S3 client
	client, err := s3.NewClient(config) // Update with your desired region
	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Delete the file or folder from the S3 bucket
	err = client.DeleteFolder(folderPath)
	if err != nil {
		response := s3.GetFailureResponse(err)
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Return a success response
	response := s3.GetUploadDeleteSuccessResponse("Folder deleted successfully")
	return c.JSON(http.StatusOK, response)
}

// ping is a simple handler to test the server
func ping(c echo.Context) error {
	response := map[string]string{"message": "pong"}
	return c.JSON(http.StatusOK, response)
}
