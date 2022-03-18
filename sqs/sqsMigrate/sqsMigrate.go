// This tool migrates all the messages from the source SQS to the destination SQS
// Usage:
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"gopkg.in/yaml.v3"
)

type AWSCredConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	SessionToken    string `yaml:"session_token"`
}

func (c AWSCredConfig) Get() credentials.StaticCredentialsProvider {
	return credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, c.SessionToken)
}

type SQSConfig struct {
	URL         string        `yaml:"url"`
	Region      string        `yaml:"region"`
	Credentials AWSCredConfig `yaml:"credentials"`
}

func (c SQSConfig) CreateClient() *sqs.Client {
	cfg := aws.Config{
		Credentials: c.Credentials.Get(),
		Region:      c.Region,
	}
	return sqs.NewFromConfig(cfg)
}

type Config struct {
	SourceSQS      SQSConfig `yaml:"source"`
	DestinationSQS SQSConfig `yaml:"destination"`
}

var (
	configPath = flag.String("conf", "config.yaml", "path to config file")
)

func main() {
	// process cmdline arg
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s: specify config file path via -conf`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// read and parse configs
	var conf Config
	ReadConfigFile(*configPath, &conf)

	// set up source SQS client
	fmt.Println(conf.SourceSQS.URL)
	fmt.Println(conf.SourceSQS.Credentials.SecretAccessKey)
	srcSQSClient := conf.SourceSQS.CreateClient()
	attr, err := srcSQSClient.GetQueueAttributes(context.Background(), &sqs.GetQueueAttributesInput{
		QueueUrl:       &conf.SourceSQS.URL,
		AttributeNames: []sqsTypes.QueueAttributeName{sqsTypes.QueueAttributeNameAll},
	})
	fmt.Println(attr)
	if err != nil {
		log.Println("Failed to get queue attributes: ", err.Error())

	}
}

// Read and parse the config yaml file given by path
func ReadConfigFile(path string, config interface{}) error {
	if path == "" {
		return fmt.Errorf("No config file path provided")
	}
	configFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Failed to open config file %s: %w", path, err)
	}
	defer configFile.Close()
	configBuffer, err := ioutil.ReadAll(configFile)
	if err != nil {
		return fmt.Errorf("Failed to read config file %s: %w", path, err)
	}
	if err = yaml.Unmarshal(configBuffer, config); err != nil {
		return fmt.Errorf("Failed to unmarshal config file %s: %w", path, err)
	}
	return nil
}
