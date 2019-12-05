// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetInstancePoolLoadBalancerAttachmentRequest wrapper for the GetInstancePoolLoadBalancerAttachment operation
type GetInstancePoolLoadBalancerAttachmentRequest struct {

	// The OCID of the instance pool.
	InstancePoolId *string `mandatory:"true" contributesTo:"path" name:"instancePoolId"`

	// The OCID of the load balancer attachment.
	InstancePoolLoadBalancerAttachmentId *string `mandatory:"true" contributesTo:"path" name:"instancePoolLoadBalancerAttachmentId"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetInstancePoolLoadBalancerAttachmentRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetInstancePoolLoadBalancerAttachmentRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetInstancePoolLoadBalancerAttachmentRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// GetInstancePoolLoadBalancerAttachmentResponse wrapper for the GetInstancePoolLoadBalancerAttachment operation
type GetInstancePoolLoadBalancerAttachmentResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The InstancePoolLoadBalancerAttachment instance
	InstancePoolLoadBalancerAttachment `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetInstancePoolLoadBalancerAttachmentResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetInstancePoolLoadBalancerAttachmentResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
