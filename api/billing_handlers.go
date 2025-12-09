package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/eniolaomotee/BlogGator-Go/internal/database"
	"github.com/stripe/stripe-go/v76"
	"io"
	"log"
	"net/http"
	"time"
)

type CreateCheckoutRequest struct {
	TierName string `json:"tier_name"`
}

type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

func (s *Server) handleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextkey).(database.User)
	var req CreateCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Get tier
	tier, err := s.db.GetTierByName(context.Background(), req.TierName)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "tier not found ")
		return
	}

	// Create or get stripe customer
	var stripeCustomerID string
	subscription, err := s.db.GetUserSubscription(context.Background(), user.ID)
	if err == nil && subscription.StripeCustomerID.Valid {
		stripeCustomerID = subscription.StripeCustomerID.String
	} else {
		customer, err := s.billing.CreateCustomer(user.Email.String, user.Name)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error creating customer")
			return
		}
		stripeCustomerID = customer.ID
	}

	// Create checkout session
	session, err := s.billing.CreateCheckoutSession(stripeCustomerID, tier.ID.String(), "https://webhook.site/233873bd-7e4b-40d8-90a8-a8448b92e32c/success", "https://webhook.site/233873bd-7e4b-40d8-90a8-a8448b92e32c/cancel")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating checkout")
		return
	}

	respondWithJson(w, http.StatusOK, CheckoutResponse{
		CheckoutURL: session.URL,
	})
}

// Get subscription
func (s *Server) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextkey).(database.User)

	subscription, err := s.db.GetUserSubscription(context.Background(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No active subscription")
		return
	}

	tier, err := s.db.GetTierByID(context.Background(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching tier")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]interface{}{
		"tier":                 tier.Name,
		"status":               subscription.Status,
		"current_period_end":   subscription.CurrentPeriodEnd,
		"cancel_at_period_end": subscription.CancelAtEndPeriod,
	})
}

// Cancel subscription
func (s *Server) handleCancelSusbscription(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextkey).(database.User)

	subscription, err := s.db.GetUserSubscription(context.Background(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No active subscription")
		return
	}
	if !subscription.StripeCustomerID.Valid {
		respondWithError(w, http.StatusBadRequest, "No stripe subscription found")
		return
	}
	// cancel in stripe
	_, err = s.billing.CancelSubscription(subscription.StripeSubscriptionID.String)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error canceling subscription")
		return
	}

	// Update in DB
	err = s.db.UpdateSubscriptionCancelation(context.Background(), database.UpdateSubscriptionCancelationParams{
		ID:                subscription.ID,
		CancelAtEndPeriod: sql.NullBool{Bool: true, Valid: true},
		UpdatedAt:         time.Now().UTC(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error updating subscription")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{
		"message": "Subscription will be canceled at period end",
	})
}

// stripe webhook handler
func (s *Server) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error reading request body")
		return
	}
	signature := r.Header.Get("Stripe-Signature")
	event, err := s.billing.HandleWebhook(payload, signature)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	// handling different event types
	switch event.Type {
	case "customer.subscription.created":
		s.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		s.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		s.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		s.handlePaymentSucceeded(event)
	case "invoice.payment_failed":
		s.handlePaymentFailed(event)
	}

	respondWithJson(w, http.StatusOK, map[string]string{"status": "success"})
}

func (s *Server) handleSubscriptionCreated(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("Error parsing webhook: %v", err)
		return
	}

	// Update database with new subscription

	// Implementation depends on your database schema
	log.Printf("Subscription created: %s", subscription.ID)
}

func (s *Server) handleSubscriptionUpdated(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("Error parsing webhook: %v", err)
		return
	}

	// Update subscription status in database
	log.Printf("Subscription updated: %s", subscription.ID)
}

func (s *Server) handleSubscriptionDeleted(event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		log.Printf("Error parsing webhook: %v", err)
		return
	}

	// Mark subscription as canceled in database
	log.Printf("Subscription deleted: %s", subscription.ID)
}

func (s *Server) handlePaymentSucceeded(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing webhook: %v", err)
		return
	}

	// Record successful payment
	log.Printf("Payment succeeded: %d cents", invoice.AmountPaid)
}

func (s *Server) handlePaymentFailed(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing webhook: %v", err)
		return
	}

	// Handle failed payment (send email, update status, etc.)
	log.Printf("Payment failed for customer: %s", invoice.Customer.ID)
}
