# Forum Project

## Overview
A simple forum for user communication and interaction. It includes features like user authentication, post categorization, likes/dislikes, and filtering. Uses SQLite for data storage and Docker for containerization.

## Features
- User registration and login with cookies for session management.
- Registered users can create posts and comments, visible to everyone.
- Like/dislike posts and comments.
- Filter posts by categories, created posts, or liked posts (logged-in users only).

## Technical Details
- Backend: SQLite database with `SELECT`, `CREATE`, and `INSERT` queries.
- Authentication: Cookies for sessions and optional password encryption.
- Frontend: Basic HTML (no frameworks allowed).
- Docker: Application containerized with Docker.
- HTTPS enabled for secure connections.

## Installation
### Prerequisites
- Docker
- Go

### Steps
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository>
   ```
2. Build and run the Docker container:
   ```bash
   docker build -t forum .
   docker run -p 8080:8080 forum
   ```
3. Access at [https://localhost:8080](https://localhost:8080).

## Project Structure
```plaintext
.
├── LICENSE
├── README.md
├── assets
│   ├── images
│   ├── static
│   ├── templates
│   └── uploads
├── cmd
│   └── main.go
├── dockerfile
├── go.mod
├── go.sum
└── internal
    ├── certs
    ├── database
    ├── loginOptions.go
    ├── models
    ├── serverRunner.go
    └── validate.go
```

## Contributions
Fork the repository and create a pull request with your changes.

## License
MIT License.