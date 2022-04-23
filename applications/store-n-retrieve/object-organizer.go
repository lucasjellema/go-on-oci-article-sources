package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

const (
	bucketName      = "go-bucket"
	compartmentOCID = "ocid1.compartment.oc1..aaaaaaaaqb4vxvxuho5h7eewd3fl6dmlh4xg5qaqmtlcmzjtpxszfc7nzbyq" // replace with the OCID of the go-on-oci compartment
	objectName      = "welcome.txt"
)

func main() {
	objectStorageClient, cerr := objectstorage.NewObjectStorageClientWithConfigurationProvider(common.DefaultConfigProvider())
	if cerr != nil {
		fmt.Printf("failed to create ObjectStorageClient : %s", cerr)
	}
	ctx := context.Background()
	namespace, cerr := getNamespace(ctx, objectStorageClient)
	if cerr != nil {
		fmt.Printf("failed to get namespace : %s", cerr)
	} else {
		fmt.Printf("Namespace : %s", namespace)
	}

	err := ensureBucketExists(ctx, objectStorageClient, namespace, bucketName, compartmentOCID)
	if err != nil {
		fmt.Printf("failed to read or create bucket : %s", err)
	}

	contentToWrite := []byte("We would like to welcome you in our humble dwellings. /n We consider it a great honor. Bla, bla.")
	objectLength := int64(len(contentToWrite))
	err = putObject(ctx, objectStorageClient, namespace, bucketName, objectName, objectLength, ioutil.NopCloser(bytes.NewReader(contentToWrite)))
	if err != nil {
		fmt.Printf("failed to write object to OCI Object storage : %s", err)
	}

	var contentRead []byte
	contentRead, err = getObject(ctx, objectStorageClient, namespace, bucketName, objectName)
	if err != nil {
		fmt.Printf("failed to get object %s from OCI Object storage : %s", objectName, err)
	}
	fmt.Printf("Object read from OCI Object Storage contains this content: %s", contentRead)

}

func getNamespace(ctx context.Context, client objectstorage.ObjectStorageClient) (string, error) {
	request := objectstorage.GetNamespaceRequest{}
	response, err := client.GetNamespace(ctx, request)
	if err != nil {
		return *response.Value, fmt.Errorf("failed to retrieve tenancy namespace : %w", err)
	}
	return *response.Value, nil
}

// bucketname needs to be unique within compartment. there is no concept of "child" buckets.
func ensureBucketExists(ctx context.Context, client objectstorage.ObjectStorageClient, namespace string, name string, compartmentOCID string) error {
	req := objectstorage.GetBucketRequest{
		NamespaceName: &namespace,
		BucketName:    &name,
	}
	// verify if bucket exists.
	response, err := client.GetBucket(ctx, req)
	if err != nil {
		if response.RawResponse.StatusCode == 404 {
			err = createBucket(ctx, client, namespace, name, compartmentOCID)
			return err
		}
		return err
	}
	fmt.Printf("bucket %s already exists", name)
	return nil
}

// bucketname needs to be unique within compartment. there is no concept of "child" buckets. using "/" separator characters in the name, the suggestion of nested bucket can be created
func createBucket(ctx context.Context, client objectstorage.ObjectStorageClient, namespace string, name string, compartmentOCID string) error {
	request := objectstorage.CreateBucketRequest{
		NamespaceName: &namespace,
	}
	request.CompartmentId = &compartmentOCID
	request.Name = &name
	request.Metadata = make(map[string]string)
	request.PublicAccessType = objectstorage.CreateBucketDetailsPublicAccessTypeNopublicaccess
	_, err := client.CreateBucket(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to create bucket on OCI : %w", err)
	} else {
		fmt.Printf("created bucket : %s", name)
	}
	return nil
}

func putObject(ctx context.Context, client objectstorage.ObjectStorageClient, namespace string, bucketName string, objectname string, contentLen int64, content io.ReadCloser) error {
	request := objectstorage.PutObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
		ObjectName:    &objectname,
		ContentLength: &contentLen,
		PutObjectBody: content,
	}
	_, err := client.PutObject(ctx, request)
	fmt.Printf("Put object %s in bucket %s", objectname, bucketName)
	if err != nil {
		return fmt.Errorf("failed to put object on OCI : %w", err)
	}
	return nil
}

func getObject(ctx context.Context, client objectstorage.ObjectStorageClient, namespace string, bucketName string, objectname string) (content []byte, err error) {
	request := objectstorage.GetObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
		ObjectName:    &objectname,
	}
	response, err := client.GetObject(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve object : %w", err)
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content from object on OCI : %w", err)
	}
	return buf.Bytes(), nil
}
