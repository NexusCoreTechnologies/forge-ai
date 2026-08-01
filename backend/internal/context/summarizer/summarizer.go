package summarizer

// Summarize returns a truncated summary preserving start of document.
func Summarize(text string, maxTokens int) string {
    if maxTokens <= 0 || len(text) <= maxTokens {
        return text
    }
    return text[:maxTokens]
}
