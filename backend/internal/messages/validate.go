package messages

// contentProblems returns field errors for message body validation.
// Shared by HTTP DTOs and the service (WebSocket entry path).
func contentProblems(content string) map[string]string {
	problems := make(map[string]string)
	if len(content) == 0 {
		problems["content"] = "required"
	} else if len(content) > 4000 {
		problems["content"] = "must be at most 4000 characters"
	}
	return problems
}
