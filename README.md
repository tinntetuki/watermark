## Watermark

Minimal Go web service, ready for Render deployment.

### Run locally

- Install Go 1.21+ (on Windows you can use `winget install GoLang.Go` or download from golang.org/dl).
- `go run .`
- Visit <http://localhost:10000>; override the port with `PORT=8080 go run .`.

### Deploy to Render

1. Commit and push this repo to GitHub/GitLab.
2. In Render, create a *Web Service* and connect the repository.
3. Set these parameters:
   - Environment: `Go`
   - Build command: `go build -o app .`
   - Start command: `./app`
4. Render automatically injects `PORT`; the server already reads it.

### Git workflow

- `git add .`
- `git commit -m "Initial Go web service"`
- `git push origin master`

