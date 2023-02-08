if archiveInTerminalState(latest) {
	msg := "Archive is in '" + *latest.ko.Status.State + "' status"
	ackcondition.SetTerminal(desired, corev1.ConditionTrue, &msg, nil)
	ackcondition.SetSynced(desired, corev1.ConditionTrue, nil, nil)
	return desired, nil
}

if archiveCreating(latest) {
	msg := "Archive is currently being created"
	ackcondition.SetSynced(desired, corev1.ConditionFalse, &msg, nil)
	return desired, requeueWaitUntilCanModify(latest)
}

if !archiveAvailable(latest) {
	msg := "Archive is not available for modification in '" +
		*latest.ko.Status.State + "' status"
	ackcondition.SetSynced(desired, corev1.ConditionFalse, &msg, nil)
	return desired, requeueWaitUntilCanModify(latest)
}
