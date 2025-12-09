package billing

import (
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

type StripeService struct {
	webhookSecret string
}

func NewStripeService(webhook string) *StripeService {
	st := &StripeService{
		webhookSecret: webhook,
	}
	return st
}

// Create a stripe customer
func (st *StripeService) CreateCustomer(email, name string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}
	return customer.New(params)
}

// Create checkout session for subscription for a user
func (st *StripeService) CreateCheckoutSession(customerID, priceID, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}
	return session.New(params)
}

// cancel subsctiption at period end
func (st *StripeService) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	return subscription.Update(subscriptionID, params)
}

// Handlewebhook to process stripe webhooks
func (st *StripeService) HandleWebhook(payload []byte, signature string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, signature, st.webhookSecret)
}
