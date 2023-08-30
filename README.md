# Forum

## Description

This project consists in creating a web forum that allows:
   * communication between users;
   * associating categories to posts;
   * liking and disliking posts and comments;
   * filtering posts.

### Docker
Docker is used to be able to containerize the project.

To build docker image run the following command:
make build

To run container for docker image run the following command:
make run

Go to: http://localhost:4000 

### Run Locally with makefile
1. make build
2. make run
3. make stop

### Run Locally without docker and makefile
Run the following command: "go run ./cmd/web/" and click on the generated URL address to go to the web page


### Authors
@aspandyar
@sfaizull


