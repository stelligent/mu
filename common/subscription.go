package common

// SubscriptionCreator for creating subscriptions
type SubscriptionCreator interface {
	CreateSubscription(topic string, protocol string, endpoint string) error
}

// SubscriptionGetter for creating subscriptions
type SubscriptionGetter interface {
	GetSubscription(topic string, protocol string, endpoint string) (interface{}, error)
}

// SubscriptionManager composite of all subscription capabilities
type SubscriptionManager interface {
	SubscriptionCreator
	SubscriptionGetter
}
