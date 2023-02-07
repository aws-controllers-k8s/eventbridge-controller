if len(r.ko.Spec.Targets) > 0 {
	if err = rm.syncTargets(
		ctx,
		r.ko.Spec.Name, r.ko.Spec.EventBusName,
		nil, r.ko.Spec.Targets,
	); err != nil {
		return nil, err
	}
}
