if err = validateRuleSpec(desired.ko.Spec); err != nil {
    return nil, ackerr.NewTerminalError(err)
}
