# chatgpt-cmd

`chatgpt-cmd` is a command line tool that allows you to interact with OpenAI's GPT models.

## Usage

To use `chatgpt-cmd`, you will need to provide your OpenAI API key(**required**) and organization ID(**omit empty**). You can also specify the program mode and proxy port.

```
Usage of ./bin/chatgpt-cmd-linux-amd64:
  -d string
        generated image dir (default "./images")
  -k string
        your openAI api key
  -m int
        program mode, 0: chat; 1: image generator
  -o string
        your organization id
  -p int
        your proxy port (default 7890)
```




### Examples

Here is an example of how to use `chatgpt-cmd` in chat mode:

> print `exit` or type `ctrl+C` to exit the program

```
./bin/chatgpt-cmd-linux-amd64 -k YOUR_API_KEY 
You: hi
>
GPT: Hello! How can I assist you today?
>>>
You: exit
```


Here is an example of how to use `chatgpt-cmd` in image generator mode:

```
./bin/chatgpt-cmd-linux-amd64 -k YOUR_API_KEY -m 1
You: A cute gengar
>
The image was saved as: images/202a865568039193e4bbd89aee3eae8e_20230331094938.png

>>>
You: exit
```

