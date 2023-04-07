package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

const (
	exitFlag = "exit" // exit flag, if you input this, the program will exit
)

var (
	apiKey         string // your openAI api key
	organizationID string // your organization id, if you don't have one, you can use the empty string

	mode      int    // program mode, 0: chat with context; 1: chat no context; 2: image generator
	proxyPort int    // your proxy port, vpn is required if you are in China
	imageDir  string // generated image dir, default is ./images

	c *openai.Client
)

func main() {
	// parse command line arguments
	flag.StringVar(&apiKey, "k", "", "your openAI api key")
	flag.StringVar(&organizationID, "o", "", "your organization id")
	flag.IntVar(&proxyPort, "p", 7890, "your proxy port")
	flag.IntVar(&mode, "m", 0, "program mode, 0: chat with context; 1: chat no context; 2: image generator")
	flag.StringVar(&imageDir, "d", "./images", "generated image dir")
	flag.Parse()
	// check arguments
	proxyURL, err := url.Parse(fmt.Sprintf("http://localhost:%d", proxyPort))
	if err != nil {
		panic(err)
	}
	// create client by config
	cfg := openai.DefaultConfig(apiKey)
	cfg.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL), // use proxy
		},
	}
	c = openai.NewClientWithConfig(cfg)
	ctx := context.Background()
	messages := make([]openai.ChatCompletionMessage, 0)
	// create a scanner to read user input
	scanner := bufio.NewScanner(os.Stdin)
	// main loop
	for {
		fmt.Print("You: ")
		scanner.Scan()
		text := scanner.Text()
		if text == exitFlag {
			break
		}
		fmt.Println(">")
		singleMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		}
		messages = append(messages, singleMsg)
		switch mode {
		case 0:
			messages, err = chatStream(ctx, messages)
			if err != nil {
				fmt.Println("chat error: ", err)
				continue
			}

		case 1:
			if _, err := chatStream(ctx, []openai.ChatCompletionMessage{singleMsg}); err != nil {
				fmt.Println("chat error: ", err)
				continue
			}

		case 2:
			if err := imageGen(ctx, text); err != nil {
				fmt.Println("generate image error: ", err)
				continue
			}
		default:
			return
		}
		fmt.Println()
		fmt.Println(">>>")
	}
}

// chatStream chat with context, if you want to chat without context, just pass a single message.
func chatStream(ctx context.Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
		Stream:   true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return messages, err
	}
	defer stream.Close()
	var builder strings.Builder
	fmt.Printf("GPT: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return messages, err
		}

		fmt.Printf(response.Choices[0].Delta.Content)
		builder.WriteString(response.Choices[0].Delta.Content)
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: builder.String(),
	})
	return messages, nil
}

// imageGen generate image with prompt, the prompt can be a sentence or a paragraph.
func imageGen(ctx context.Context, prompt string) error {
	reqBase64 := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := c.CreateImage(ctx, reqBase64)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return nil
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return nil
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		fmt.Printf("PNG decode error: %v\n", err)
		return nil
	}
	if !pathExists(imageDir) {
		if err := os.MkdirAll(imageDir, 0777); err != nil {
			return err
		}
	}
	filename := fmt.Sprintf("%s_%s.png", md5Str(prompt), time.Now().Format("20060102150405"))
	filename = path.Join(imageDir, filename)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("File creation error: %v\n", err)
		return nil
	}
	defer file.Close()

	if err := png.Encode(file, imgData); err != nil {
		fmt.Printf("PNG encode error: %v\n", err)
		return nil
	}

	fmt.Println("The image was saved as:", filename)
	return nil
}

// md5Str generate md5 string, the input string will be converted to []byte.
func md5Str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// pathExists check if the file or directory exists, if exists, return true, otherwise return false.
func pathExists(p string) bool {
	// Check if the file or directory exists
	_, err := os.Stat(p)
	return !errors.Is(err, fs.ErrNotExist)
}
