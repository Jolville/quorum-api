package communications

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
)

type SRVCommunications interface {
	SendEmail(request SendEmailRequest) error
}

type SendEmailRequest struct {
	ToEmail    string
	TemplateID int
	Subject    string
	Variables  map[string]interface{}
}

type srvCommunications struct {
	mjClient *mailjet.Client
}

func New(mjApiKeyPublic string, mjApiKeyPrivate string) SRVCommunications {
	mjClient := mailjet.NewMailjetClient(
		mjApiKeyPublic, mjApiKeyPrivate,
	)
	return &srvCommunications{
		mjClient: mjClient,
	}
}

func (s *srvCommunications) SendEmail(request SendEmailRequest) error {
	if os.Getenv("GO_ENV") == "local" {
		fmt.Printf("Email subject %q", request.Subject)
		spew.Dump(request.Variables)
		return nil
	}
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "noreply@quorumvote.com",
				Name:  "Verify your Quorum Account",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: request.ToEmail,
				},
			},
			TemplateID:       request.TemplateID,
			TemplateLanguage: true,
			Subject:          request.Subject,
			Variables:        request.Variables,
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	res, err := s.mjClient.SendMailV31(&messages)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	fmt.Printf("email result: %v", res)
	return nil
}
