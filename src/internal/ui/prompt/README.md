# prompt package
This is for the prompt modal of t4gfm

Handles user input updates, gfm model updates, and returns a PromptAction to model. 


# Coverage

```bash
cd /path/to/ui/prompt
# Basic coverage
go test -cover

# HTML report
go test -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
```
Current coverage is 91.3%.
