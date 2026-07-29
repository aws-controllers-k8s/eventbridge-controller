if err := rm.setResourceAdditionalFields(ctx, ko); err != nil {
    return nil, err
}
if resp.Policy != nil {
    ko.Spec.Policy = resp.Policy
} else {
    ko.Spec.Policy = nil
}
