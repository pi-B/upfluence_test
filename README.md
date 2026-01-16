# Uplfuence analysis-api

This document aims at centralizing all the informations regarding this piece of sofwtare.

# Introduction

The API aims at providing an analysis report on the social media posts that where processed by Upfluence and published on its SSE stream at https://stream.upfluence.co/stream.

The user can contact the API only on the `GET /analysis` route, having to provide two parameters in its query : 
- `duration`: the duration for the analysis to run, using a format that can be parsed and transformed into a duration e.g `20s` for a 20 second analysis, `2m` for 2 minutes, `4h` for 4 hours.
- `dimension`: the type of stat the user wants to analyse over every posts processed by Upfluence. It can only be 
    - `pin`
    - `instagram_media`
    - `youtube_video`
    - `article`
    - `tweet`
    - `facebook_status`

The API is currently configured to run behind the port `8080`, any change to this should be repercuted in the automatisation files (`Dockerfile`, `docker-compose.yml`).

## Technical choices

### API framework

The API is only using the Go standard library to operate, except for its HTTP server and request handling, in order to be quicker to write.

I have selected the `Gin` framework to perform these tasks, Gin being the go-to tool in the Go community to set up API quickly. It prevented me from spending too much time writing a lot of boilerplate code and ship a solution quickly. The tradeoff for this being the bloatload Gin brings to the project, in the `go.mod` we can see the only imported package requires 20+ other packages.

I had never used Gin previously, working only with `beego`, which is an interesting choice for its rout-creation-over-comments and swagger generation features, but that would be a bit too overkill for this particuliar project. 

### Architecture

The project uses a simple API architecture where the router is located in the main file for simplicity concerns, then controllers, services and models are separated in order to make the code more readable and keep class/methods short. 

The controller is here to make sure the received queries are well formatted, i.e contain the required parameters with appropriate formats. They then used the `analysis` service to contact the Upfluence stream and retrieve the data it returns to the client.

All the "business" logic is located in the only service, althought different method are handling the different steps of the data retrieving process. The service relies on the `watcher` class to handle the useful data it extracted and store it.

Two types of models are used here : 
- `watcher` : it is instanciated to store the nature of the client's request throught the `TargetDimension` attribute and to store the retrieved data that fit the client's request.
- `SocialsData` : it uses the tag feature of Go to store informations that are found in any post (`id` and `timestamp`) and retrieve transparently any of the dimension we are interested in (likes, comments, favorites, retweets) when a JSON structured object is unmarshalled in it.

The Analysis and Watcher communicate through a channel were each message is parsed and fed to a new goroutine in charge of processing the post's data. With this way of doing we make sure that the processing is not blocking, allowing the service to process a fast input of events.  

## Critical analysis : 

In this part I will try to analyse the weaknesses of the delivered code.

### Improvements 

Using the Gin package to handle the HTTP server and requests comes with an increased binary size. While the project is quite small and only consists of one route it imports a lot of dependencies. The resulting image, while embarquing a minimal environment using `SCRATCH`, is still almost 10 Mb. Using this environment will also reduce the possibilities to interact with the container's environment, notably lacking a shell like `bash` or `sh`.

The solution would suffer from any downtime, even the shortest, in the stream provided by upfluence to retrieve the data. Should the stream cut I have not implemented a way for the solution to detect it and wait for the stream to be available again. I guess this would result in an Internal Error and `500` answer. The solution is also not capable of detecting and logging this event **precisely**.

It also lacks a proper detection and logging of any malformated data. I had thought the `logger` class could implement a method to store the processing errors and add this information as another report in the final response should they exist.

Finaly, I have observed, using the small test in `tests/test_carge_increase.go`, that when a very large number of requests are handled at the same time (above 100) the API was suffering problems to establish a TLS connetion to the uplfuence stream. This might be handled with the buffering of the request.

### Tests

Unfortunately I have never used the testing module of Go. I did not want to spend a lot of time learning how to use, for the final implementation being of poor quality, due to my inexperience.
However, I have used progressive testing while writing the code, starting from the implementation of the models, then the service and finally the controller. I have also generated data to test the solution in the `tests` directory.

## Deployment

The solution can be deployed either directly on the host machine 

```
(inside the cloned directory)
go run .

or 

go build . && ./analysis-api
```

or in a docker container 

```
(inside the cloned directory)
docker build . -t analysis-api
docker compose up (-d if you don't want to logs to appear in your terminal)
```

