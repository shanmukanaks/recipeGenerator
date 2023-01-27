package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Clarifai/clarifai-go-grpc/proto/clarifai/api"
	"github.com/Clarifai/clarifai-go-grpc/proto/clarifai/api/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func recipeGen(ingredients string) string {
	// Set up the request
	url := "https://api.openai.com/v1/engines/text-davinci-003/generate"
	apiKey := "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"
	prompt := ingredients
	var jsonStr = []byte(`{
			"context":"`+prompt+`",
			"length":1000,
			"top_p":0.5,
			"temperature":1,
			"best_of":1,
			"completions":1,
			"stream":false,
			"logprobs":0,
			"stop":"###"
	}`)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
			fmt.Println(err)
			return err.Error()
	}
	defer resp.Body.Close()

	// Read the response
	body, _ := ioutil.ReadAll(resp.Body)
	//Convert the byte array to string
	bodyStr := string(body)
	//split the string into an array of strings
	bodyArr := strings.Split(bodyStr, "\n")

	fs := "Recipe"

	for i := 0; i < len(bodyArr)/2; i++ {
		for j := 0; j < len(bodyArr[i]); j++ {
			// fmt.Println(string(bodyArr[i][j]))
			if bodyArr[i][j] == '<' {
				break
			}
			if bodyArr[i][j] != '"' && bodyArr[i][j] != ',' && bodyArr[i][j] != '[' && bodyArr[i][j] != ']' && bodyArr[i][j] != '{' && bodyArr[i][j] != '}' {
				fs += fmt.Sprintf(string(bodyArr[i][j]))
			}
		}
	}

	i := strings.Index(fs, "context:0completion:0tokens")
	recipe := fs[:i]
	recipe = strings.Split(recipe, "text-davinci-003data:object:snippettext:")[1]
	return recipe
}

func main() {
	conn, err := grpc.Dial(
			"api.clarifai.com:443",
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
	)
	if err != nil {
			panic(err)
	}
	client := api.NewV2Client(conn)

	ctx := metadata.AppendToOutgoingContext(
			context.Background(),
			"Authorization", "Key 4ad0e68afae14b40af2e78fff189a9e1",
	)
	// This is a publicly available model ID.
	var GeneralModelId = "food-item-recognition"
	response, err := client.PostModelOutputs(
			ctx,
			&api.PostModelOutputsRequest{
					ModelId: GeneralModelId,
					Inputs: []*api.Input{
							{
									Data: &api.Data{
											Image: &api.Image{
													Url: "https://www.allrecipes.com/thmb/hSLLdeWkvzR50-YXlUaWOzoh5Ck=/1500x0/filters:no_upscale():max_bytes(150000):strip_icc()/20513-classic-waffles-mfs-014_step1-879a0c96dd8b4f1095828445726351c6.jpg",
											},
									},
							},
					},
			},
	)
	if err != nil {
			panic(err)
	}
	if response.Status.Code != status.StatusCode_SUCCESS {
			panic(fmt.Sprintf("Failed response: %s", response))
	}

	food := "What are some recipes for making a dish with"

	for _, concept := range response.Outputs[0].Data.Concepts {
		food += " " + concept.Name
	}

	rec := recipeGen(food)

	fmt.Println("food: ", food)
	fmt.Println("")
	fmt.Println(rec)
}