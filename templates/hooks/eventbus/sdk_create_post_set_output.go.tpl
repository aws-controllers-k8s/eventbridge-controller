if ko.Spec.Policy != nil && *ko.Spec.Policy != "" {
    if err := rm.putPolicy(ctx, ko); err != nil {
        return nil, err
    }
}
