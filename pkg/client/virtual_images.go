// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	consts "github.com/hpe-hcss/vmaas-cmp-go-sdk/pkg/common"
	models "github.com/hpe-hcss/vmaas-cmp-go-sdk/pkg/models"
)

var (
	_ context.Context
)

type VirtualImagesApiService struct {
	Client APIClientHandler
	Cfg    Configuration
}

/*
VirtualImageApiService
Get All Virtual images
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param serviceInstanceId
 * @param name/phrase optional
@return models.VirtualImages
*/
func (a *VirtualImagesApiService) GetAllVirtualImages(ctx context.Context, param map[string]string) (models.VirtualImages, error) {
	var (
		localVarHttpMethod   = strings.ToUpper("Get")
		localVarPostBody     interface{}
		localVarFileName     string
		localVarFileBytes    []byte
		virtualImageResponse models.VirtualImages
	)

	// create path and map variables
	localVarPath := fmt.Sprintf("%s/%s/%s", a.Cfg.Host, consts.VmaasCmpApiBasePath,
		consts.VirtualImagePath)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := getUrlValues(param)
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.Client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return virtualImageResponse, err
	}

	localVarHttpResponse, err := a.Client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return virtualImageResponse, err
	}

	if localVarHttpResponse.StatusCode >= 300 {
		return virtualImageResponse, ParseError(localVarHttpResponse)
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	defer localVarHttpResponse.Body.Close()
	if err != nil {
		return virtualImageResponse, err
	}

	if err := json.Unmarshal(localVarBody, &virtualImageResponse); err != nil {
		return virtualImageResponse, err
	}
	return virtualImageResponse, nil
}
