package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/uuid"
)

func main() {
	key := os.Args[1]
	secret := os.Args[2]
	region := os.Args[3]

	client := getSecretClient(region, key, secret)
	secretList := getList(client)

	fileCSV := createOutputFile()
	defer fileCSV.Close()

	writerCSV := csv.NewWriter(fileCSV)
	defer writerCSV.Flush()

	secretsToCSV(secretList, client, writerCSV)

	fmt.Printf("file %s created.\n", fileCSV.Name())
}

func createOutputFile() *os.File {
	runID := fmt.Sprintf("%s_%s.csv", time.Now().Format("20060102150405"), uuid.New().String()[:8])
	fileCSV, err := os.Create(runID)
	if err != nil {
		log.Fatal(err)
	}
	return fileCSV
}

func getSecretClient(region string, key string, secret string) *secretsmanager.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")),
	)
	if err != nil {
		log.Fatal(err)
	}
	return secretsmanager.NewFromConfig(cfg)
}

func secretsToCSV(secretList *secretsmanager.ListSecretsOutput, client *secretsmanager.Client, writerCSV *csv.Writer) {
	var secretValue *secretsmanager.GetSecretValueOutput
	var err error
	writerCSV.Write([]string{"key", "value", "desc"}) // header of the CSV file
	for _, object := range secretList.SecretList {
		secretParams := &secretsmanager.GetSecretValueInput{
			SecretId:     object.Name,
			VersionStage: aws.String("AWSCURRENT"),
		}
		secretValue, err = client.GetSecretValue(context.TODO(), secretParams)
		if err != nil {
			log.Fatal(err.Error())
		}
		writerCSV.Write([]string{
			aws.ToString(object.Name),        // key,
			*secretValue.SecretString,        // value
			aws.ToString(object.Description), // description
		})
	}
}

func getList(client *secretsmanager.Client) *secretsmanager.ListSecretsOutput {
	listSecretsOutput, err := client.ListSecrets(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return listSecretsOutput
}
