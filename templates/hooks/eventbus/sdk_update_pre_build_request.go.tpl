if delta.DifferentAt("Spec.Policy") {
    if err := rm.syncPolicy(ctx, latest, desired); err != nil {
        return nil, err
    }
}
