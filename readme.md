### Project README

This project is a simple HTTP/3 server built in Go that handles URL redirection based on stored redirect records in a Redis database. Below are detailed instructions on how to set up and use the project.

---

### Overview

The project consists of an HTTP/3 server implemented using the `quic-go/http3` package. It listens for incoming requests, retrieves the corresponding redirect record from a Redis database, and redirects the client to the target URL if a matching record is found.

### Setup and Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/your_username/your_project.git
   cd your_project
   ```

2. **Install dependencies:**

   Ensure you have Go 1.16+ installed on your machine. Use the following command to install the required dependencies:

   ```bash
   make dependencies
   ```

   This command will ensure that all necessary Go dependencies are installed.

### Configuration

- Ensure that a Redis server is running locally on the default port (6379).

### Usage

1. **Start the server:**

   Run the following command to start the HTTP/3 server:

   ```bash
   make run
   ```

   This command will start the server and listen on port 443. You can then access the server using a web browser or an HTTP client.

2. **Update redirect records:**

   To update redirect records, send a POST request to the `/api/update-redirect` endpoint with the appropriate JSON payload containing the redirect entries.

   Example:

   ```bash
   curl -X POST -H "Content-Type: application/json" -d '{"from": "old_url", "to": "new_url"}' http://localhost:443/api/update-redirect
   ```

3. **Get all redirect records:**

   To retrieve all redirect records, send a GET request to the `/api/get-redirects` endpoint.

   Example:

   ```bash
   curl http://localhost:443/api/get-redirects
   ```

### Contributing

If you would like to contribute to the project, follow these steps:

1. Fork the project on GitHub.
2. Create a new branch with a descriptive name.
3. Make your changes and commit them with clear messages.
4. Push your changes to your fork.
5. Submit a pull request to the main repository's `master` branch.

Project Link: [https://github.com/your_username/your_project](https://github.com/your_username/your_project)
