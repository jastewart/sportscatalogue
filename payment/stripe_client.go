package payment

import (
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/caleb/sports-catalog/models"
)

type StripeClient struct {
	SuccessURL string
	CancelURL  string
	WebhookKey string
}

func NewStripeClient(apiKey, successURL, cancelURL, webhookKey string) *StripeClient {
	stripe.Key = apiKey // safe to set globally: one process, one merchant account
	return &StripeClient{SuccessURL: successURL, CancelURL: cancelURL, WebhookKey: webhookKey}
}

// CreateCheckoutSession builds a Stripe Checkout Session for the given order
// items. In test mode with a Stripe test secret key, this produces a fully
// working mock checkout — card 4242 4242 4242 4242 will "pay" successfully,
// so you don't need real product/payment infrastructure to exercise the flow.
func (c *StripeClient) CreateCheckoutSession(items []models.OrderItem, customerEmail string) (*stripe.CheckoutSession, error) {
	lineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(items))
	for _, it := range items {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			Quantity: stripe.Int64(int64(it.Quantity)),
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency:   stripe.String("usd"),
				UnitAmount: stripe.Int64(it.UnitPriceCents),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(it.Name),
				},
			},
		})
	}

	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems:         lineItems,
		SuccessURL:        stripe.String(c.SuccessURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:         stripe.String(c.CancelURL),
		CustomerEmail:      stripe.String(customerEmail),
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}
	return sess, nil
}

// ConstructWebhookEvent verifies the Stripe-Signature header and decodes the
// event body. Always verify signatures — never trust an unauthenticated
// POST claiming "payment succeeded".
func (c *StripeClient) ConstructWebhookEvent(payload []byte, sigHeader string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, sigHeader, c.WebhookKey)
}
