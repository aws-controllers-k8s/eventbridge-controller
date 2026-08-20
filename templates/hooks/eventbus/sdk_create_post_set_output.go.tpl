if ko.Spec.Policy != nil && *ko.Spec.Policy != "" {
    // The resource-based policy is not part of the CreateEventBus API; it is
    // applied via a separate PutPermission call, which lives in the update
    // path (customUpdate -> syncPolicy). Calling PutPermission here would
    // conflate a policy failure with a failure to create the event bus.
    // Instead, mark the resource as not-yet-synced so the reconciler requeues
    // and enters the update path to reconcile the policy.
    msg := "EventBus policy is pending, requeuing to reconcile via update path"
    ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, &msg, nil)
}
