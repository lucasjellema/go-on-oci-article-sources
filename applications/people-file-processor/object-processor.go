package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

const (
	RUN_WITH_INSTANCE_PRINCIPAL_AUTHENTICATION = false
)

func RetrieveObject(objectName string, bucketName string, compartmentOCID string) ([]byte, error) {
	// for running in an environment (such as an OCI Compute Instance or OKE cluster) that inherits instance principal authentication
	var objectStorageClient objectstorage.ObjectStorageClient
	if RUN_WITH_INSTANCE_PRINCIPAL_AUTHENTICATION {
		configurationProvider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			log.Printf("failed to get oci configurationprovider based on instance principal authentication : %s", err)
			return nil, err
		}
		objectStorageClient, err = objectstorage.NewObjectStorageClientWithConfigurationProvider(configurationProvider)
		if err != nil {
			log.Printf("failed to create ObjectStorageClient : %s", err)
			return nil, err
		}
	} else {
		// for running in an environment with ~/.oci/config in place:
		var err error
		objectStorageClient, err = objectstorage.NewObjectStorageClientWithConfigurationProvider(common.DefaultConfigProvider())

		if err != nil {
			log.Printf("failed to create ObjectStorageClient : %s", err)
			return nil, err
		}
	}
	ctx := context.Background()
	namespace, cerr := getNamespace(ctx, objectStorageClient)
	if cerr != nil {
		log.Printf("failed to get namespace : %s", cerr)
	} else {
		log.Printf("Namespace : %s", namespace)
	}

	var contentRead []byte
	contentRead, err := getObject(ctx, objectStorageClient, namespace, bucketName, objectName)
	if err != nil {
		log.Printf("failed to get object %s from OCI Object storage : %s", objectName, err)
		return nil, err
	}
	return contentRead, nil
}

func getNamespace(ctx context.Context, client objectstorage.ObjectStorageClient) (string, error) {
	request := objectstorage.GetNamespaceRequest{}
	response, err := client.GetNamespace(ctx, request)
	if err != nil {
		return *response.Value, fmt.Errorf("failed to retrieve tenancy namespace : %w", err)
	}
	return *response.Value, nil
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
