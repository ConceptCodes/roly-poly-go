# Roly Poly

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Rest Api to manage polls 🤷🏾‍♂️

## Features
- Authorization by Api Key

## Prerequisites

[![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/doc/install) (version 1.21 or higher)

## Installation

1. Clone this repository:

   ```sh
   git clone https://github.com/conceptcodes/roly-poly-go.git
   cd roly-poly-go
   ```

## Usage

1. Run the server

  ```sh
  gow run cmd/server/main.go
  ```

2. Verify that the service is running

  ```sh
  curl http://localhost:8080/health/alive
  ```
  ```json
  {
    "message": "Service is alive",
    "data": null,
    "error_code": ""
  }
  ```

3. Onboard a new tenant. This action will return your api key, which you will use to authenticate all other requests.

  ```sh
  curl --location 'http://localhost:8080/api/onboard' \
  --header 'Content-Type: application/json' \
  --data '{
      "first_name": "sample_first_name",
      "last_name": "sample_last_name"
  }'
  ```
  ```json
  {
    "message": "User onboarded successfully",
    "data": {
        "api_key": "sample_api_key",
        "first_name": "sample_first_name",
        "last_name": "sample_last_name"
    },
    "error_code": ""
  }
  ```

4. Create a new Poll

  ```sh
  ```
  ```json
  ```

## Roadmap

- [ ] Add known errors to the api
- [ ] Add an endpoint to generate reports
- [ ] Integrate with a OpenAi to offer a chatbot over results
- [ ] Add more tests