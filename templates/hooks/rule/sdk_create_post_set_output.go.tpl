if len(ko.Spec.Targets) > 0 {
	if err = rm.syncTargets(
	    ctx,
		ko.Spec.Name, ko.Spec.EventBusName,
		ko.Spec.Targets, nil,
	); err != nil {
		return nil, err
	}
}
