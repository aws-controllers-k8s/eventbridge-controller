if err = validateRuleSpec(desired.ko.Spec); err != nil {
		return nil, ackerr.NewTerminalError(err)
}
if delta.DifferentAt("Spec.Tags") {
	if err = rm.syncTags(ctx, desired, latest); err != nil {
		return nil, err
	}
}
if delta.DifferentAt("Spec.Targets") {
	if err = rm.syncTargets(
		ctx,
		desired.ko.Spec.Name, desired.ko.Spec.EventBusName,
		desired.ko.Spec.Targets, latest.ko.Spec.Targets,
	); err != nil {
		return nil, err
	}
}
if !delta.DifferentExcept("Spec.Tags", "Spec.Targets") {
	return desired, nil
}
