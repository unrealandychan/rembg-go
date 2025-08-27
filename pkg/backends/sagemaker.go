package backends

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"
)

// InvokeSageMaker calls a SageMaker real-time endpoint with raw image bytes and returns the raw response bytes.
func InvokeSageMaker(ctx context.Context, endpoint string, payload []byte) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := sagemakerruntime.NewFromConfig(cfg)

	input := &sagemakerruntime.InvokeEndpointInput{
		Body:         payload,
		EndpointName: aws.String(endpoint),
		ContentType:  aws.String("application/octet-stream"),
	}
	resp, err := client.InvokeEndpoint(ctx, input)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
