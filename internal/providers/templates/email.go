package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

const (
	// Base email template with proper styling and unsubscribe mechanism
	BaseEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Subject}}</title>
    <style>
        body {
            font-family: 'Arial', sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .email-container {
            background-color: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            border-bottom: 2px solid #007bff;
            padding-bottom: 20px;
            margin-bottom: 30px;
        }
        .header h1 {
            color: #007bff;
            margin: 0;
        }
        .content {
            margin-bottom: 30px;
        }
        .content h2 {
            color: #333;
            margin-top: 0;
        }
        .content p {
            margin-bottom: 15px;
        }
        .footer {
            border-top: 1px solid #ddd;
            padding-top: 20px;
            font-size: 12px;
            color: #666;
            text-align: center;
        }
        .unsubscribe-link {
            color: #666;
            text-decoration: none;
        }
        .unsubscribe-link:hover {
            text-decoration: underline;
        }
        .topic-tag {
            background-color: #007bff;
            color: white;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: bold;
            text-transform: uppercase;
            margin-right: 5px;
        }
        @media (max-width: 480px) {
            body {
                padding: 10px;
            }
            .email-container {
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <h1>Newsletter</h1>
            {{if .TopicName}}
            <span class="topic-tag">{{.TopicName}}</span>
            {{end}}
        </div>
        
        <div class="content">
            <h2>{{.Subject}}</h2>
            <div>
                {{.Body}}
            </div>
        </div>
        
        <div class="footer">
            <p>You received this email because you subscribed to our newsletter.</p>
            {{if .UnsubscribeURL}}
            <p>
                <a href="{{.UnsubscribeURL}}" class="unsubscribe-link">
                    Unsubscribe from this newsletter
                </a>
            </p>
            {{end}}
            <p>© 2025 Newsletter Service. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	// Legacy template for backward compatibility
	EmailTemplate = BaseEmailTemplate

	// Plain text fallback template
	PlainTextTemplate = `
{{.Subject}}

{{.Body}}

---
You received this email because you subscribed to our newsletter.
{{if .UnsubscribeURL}}
To unsubscribe, visit: {{.UnsubscribeURL}}
{{end}}

© 2025 Newsletter Service. All rights reserved.
`
)

type EmailTemplateData struct {
	Subject        string
	Body           template.HTML
	TopicName      string
	UnsubscribeURL string
	SubscriberID   uint
	ContentID      uint
}

// Legacy EmailData for backward compatibility
type EmailData struct {
	Subject string
	Body    template.HTML
}

// GenerateEmailHTML generates a styled HTML email from template data
func GenerateEmailHTMLWithData(data EmailTemplateData) (string, error) {
	tmpl, err := template.New("email").Parse(BaseEmailTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	// Convert plain text body to HTML if needed
	if !strings.Contains(string(data.Body), "<") {
		data.Body = template.HTML(convertToHTMLParagraphs(string(data.Body)))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return buf.String(), nil
}

// GenerateEmailHTML generates a styled HTML email (backward compatibility)
func GenerateEmailHTML(subject, body string) (string, error) {
	data := EmailTemplateData{
		Subject: subject,
		Body:    template.HTML(convertToHTMLParagraphs(body)),
	}
	return GenerateEmailHTMLWithData(data)
}

// GenerateEmailHTMLWithUnsubscribe generates HTML email with unsubscribe link
func GenerateEmailHTMLWithUnsubscribe(data EmailTemplateData, baseURL string) (string, error) {
	if baseURL != "" && data.SubscriberID > 0 && data.ContentID > 0 {
		data.UnsubscribeURL = fmt.Sprintf("%s/unsubscribe?subscriber=%d&content=%d",
			strings.TrimRight(baseURL, "/"), data.SubscriberID, data.ContentID)
	}

	return GenerateEmailHTMLWithData(data)
}

// GenerateEmailText generates plain text email
func GenerateEmailText(data EmailTemplateData) (string, error) {
	tmpl, err := template.New("email-text").Parse(PlainTextTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return buf.String(), nil
}

// convertToHTMLParagraphs converts plain text with line breaks to HTML paragraphs
func convertToHTMLParagraphs(text string) string {
	// Replace double newlines with paragraph breaks
	paragraphs := strings.Split(text, "\n\n")

	var htmlParagraphs []string
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			// Replace single newlines within paragraphs with <br>
			p = strings.ReplaceAll(p, "\n", "<br>")
			htmlParagraphs = append(htmlParagraphs, fmt.Sprintf("<p>%s</p>", strings.TrimSpace(p)))
		}
	}

	return strings.Join(htmlParagraphs, "\n")
}

// ValidateTemplateData validates that required template data is provided
func ValidateTemplateData(data EmailTemplateData) error {
	if data.Subject == "" {
		return fmt.Errorf("email subject is required")
	}
	if string(data.Body) == "" {
		return fmt.Errorf("email body is required")
	}
	return nil
}
