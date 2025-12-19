package sdk

import "errors"

var (
	// ErrInvalidCapabilityManifest indicates the manifest was malformed or missing required fields.
	ErrInvalidCapabilityManifest = errors.New("sdk: invalid capability manifest")
	// ErrLifecycleViolation is returned when lifecycle hooks are invoked out of order.
	ErrLifecycleViolation = errors.New("sdk: lifecycle order violation")
	// ErrMissingIdentity indicates required identity fields were not provided.
	ErrMissingIdentity = errors.New("sdk: missing agent identity")
	// ErrNoPublisher indicates a publish was attempted without a configured publisher.
	ErrNoPublisher = errors.New("sdk: publisher not configured")
)
