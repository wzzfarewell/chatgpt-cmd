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
	"time"

	"github.com/sashabaranov/go-openai"
)

const (
	exitFlag = "exit"
)

var (
	apiKey         string
	organizationID string

	mode      int
	proxyPort int
	imageDir  string

	c *openai.Client
)

func main() {
	flag.StringVar(&apiKey, "k", "", "your openAI api key")
	flag.StringVar(&organizationID, "o", "", "your organization id")
	flag.IntVar(&proxyPort, "p", 7890, "your proxy port")
	flag.IntVar(&mode, "m", 0, "program mode, 0: chat; 1: image generator")
	flag.StringVar(&imageDir, "d", "./images", "generated image dir")
	flag.Parse()
	proxyURL, err := url.Parse(fmt.Sprintf("http://localhost:%d", proxyPort))
	if err != nil {
		panic(err)
	}
	cfg := openai.DefaultConfig(apiKey)
	cfg.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	c = openai.NewClientWithConfig(cfg)
	ctx := context.Background()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		scanner.Scan()
		text := scanner.Text()
		if text == exitFlag {
			break
		}
		fmt.Println(">")
		switch mode {
		case 0:
			if err := chatStream(ctx, text); err != nil {
				panic(err)
			}

		case 1:
			if err := imageGen(ctx, text); err != nil {
				panic(err)
			}
		default:
			return
		}
		fmt.Println()
		fmt.Println(">>>")
	}
}

func chatStream(ctx context.Context, content string) error {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		// MaxTokens: 2000,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
		Stream: true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return err
	}
	defer stream.Close()

	fmt.Printf("GPT: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return err
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
}

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

func md5Str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func pathExists(p string) bool {
	// Check if the file or directory exists
	_, err := os.Stat(p)
	return !errors.Is(err, fs.ErrNotExist)
}
