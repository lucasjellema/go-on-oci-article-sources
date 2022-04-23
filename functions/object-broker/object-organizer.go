package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/oracle/oci-go-sdk/v54/common/auth"
	"github.com/oracle/oci-go-sdk/v54/objectstorage"
)

func CreateObject(objectName string, bucketName string, compartmentOCID string) (string, error) {
	configurationProvider, err := auth.ResourcePrincipalConfigurationProvider()
	if err != nil {
		log.Printf("failed to get oci configurationprovider based on resource principal authentication : %s", err)
		return "", err
	}
	objectStorageClient, cerr := objectstorage.NewObjectStorageClientWithConfigurationProvider(configurationProvider)
	if cerr != nil {
		log.Printf("failed to create ObjectStorageClient : %s", cerr)
		return "", err
	}
	ctx := context.Background()
	namespace, cerr := getNamespace(ctx, objectStorageClient)
	if cerr != nil {
		log.Printf("failed to get namespace : %s", cerr)
	} else {
		log.Printf("Namespace : %s", namespace)
	}

	err = ensureBucketExists(ctx, objectStorageClient, namespace, bucketName, compartmentOCID)
	if err != nil {
		log.Printf("failed to read or create bucket : %s", err)
		return "", err
	}

	contentToWrite := []byte("We would like to welcome you in our humble dwellings. /n We consider it a great honor. Bla, bla.")
	objectLength := int64(len(contentToWrite))
	err = putObject(ctx, objectStorageClient, namespace, bucketName, objectName, objectLength, ioutil.NopCloser(bytes.NewReader(contentToWrite)))
	if err != nil {
		log.Printf("failed to write object to OCI Object storage : %s", err)
		return "", err
	}

	var contentRead []byte
	contentRead, err = getObject(ctx, objectStorageClient, namespace, bucketName, objectName)
	if err != nil {
		log.Printf("failed to get object %s from OCI Object storage : %s", objectName, err)
		return "", err
	}
	log.Printf("Object read from OCI Object Storage contains this content: %s", contentRead)
	return fmt.Sprintf("Object %s written to bucket %s and then read back from OCI Object Storage with this content: %s", objectName, bucketName, contentRead), nil
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
func createBucket(ctx context.Context, client objectstorage.ObjectStorageClient, namespace string, bucketName string, compartmentOCID string) error {
	request := objectstorage.CreateBucketRequest{
		NamespaceName: &namespace,
	}
	request.CompartmentId = &compartmentOCID
	request.Name = &bucketName
	request.Metadata = make(map[string]string)
	request.PublicAccessType = objectstorage.CreateBucketDetailsPublicAccessTypeNopublicaccess
	_, err := client.CreateBucket(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to create bucket on OCI : %w", err)
	} else {
		fmt.Printf("created bucket : %s", bucketName)
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
