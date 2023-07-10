package s3

import (
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
)

func GetFailureResponse(err error) FailureResponse {
	return FailureResponse{
		Status:       "Failure",
		ResponseCode: http.StatusInternalServerError,
		ErrorMessage: err.Error(),
	}
}

func GetListFolderSuccessResponse(payload *ListFilesResponse) SuccessResponse {
	return SuccessResponse{
		Status:       "Success",
		ResponseCode: http.StatusOK,
		Data:         payload,
	}
}

func GetUploadDeleteSuccessResponse(filePath string) SuccessResponse {
	return SuccessResponse{
		Status:       "Success",
		ResponseCode: http.StatusOK,
		Data:         filePath,
	}
}

// custom function to sort the files by name or last modified
func SortFiles(files []ObjectDetails, c echo.Context) []ObjectDetails {
	sortBy := c.QueryParam("sortBy")
	sortOrder := c.QueryParam("sortOrder")

	if sortBy == "" {
		sortBy = "name"
	}

	if sortOrder == "" {
		sortOrder = "asc"
	}

	if sortBy == "name" {
		if sortOrder == "asc" {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Name < files[j].Name
			})
		} else {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Name > files[j].Name
			})
		}
	} else if sortBy == "date" {
		if sortOrder == "asc" {
			sort.Slice(files, func(i, j int) bool {
				return files[i].LastModified.Before(files[j].LastModified)
			})
		} else {
			sort.Slice(files, func(i, j int) bool {
				return files[i].LastModified.After(files[j].LastModified)
			})
		}
	} else if sortBy == "type" {
		if sortOrder == "asc" {
			sort.Slice(files, func(i, j int) bool {
				return files[i].IsFolder
			})
		} else {
			sort.Slice(files, func(i, j int) bool {
				return !files[i].IsFolder
			})
		}
	} else if sortBy == "size" {
		if sortOrder == "asc" {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Size < files[j].Size
			})
		} else {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Size > files[j].Size
			})
		}
	}

	return files
}
