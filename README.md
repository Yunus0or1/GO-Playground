# GO-Programming

Playground to play football against GO. Exploring different aspects.


# To build a GO program
    ```
    set GOOS=linux 
    set GOARCH=amd64 
    set CGO_ENABLED=0
    go build -o main scheduler_aws_lambda.go
    zip main.zip main
    ```
